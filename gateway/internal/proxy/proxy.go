package proxy

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/eagi/gateway/internal/process"
	"github.com/eagi/gateway/internal/router"
)

type originalIDMapping struct {
	SessionID  string
	OriginalID interface{}
}

type RemoteDomain struct {
	Name    string
	SSEURL  string
	PostURL string
}

// Proxy handles HTTP/SSE MCP requests and proxies them to stdio Node processes or remote SSE servers
type Proxy struct {
	manager       *process.Manager
	Router        *router.Router
	clients       map[string]chan []byte // SSE connections
	DefaultDomain string
	mu            sync.Mutex

	transactions map[string]originalIDMapping
	txMu         sync.Mutex

	subscriptions map[string]map[string]bool // URI -> Set of SessionIDs
	subMu         sync.Mutex

	remoteDomains map[string]*RemoteDomain
	remoteMu      sync.RWMutex
}

func NewProxy(manager *process.Manager) *Proxy {
	return &Proxy{
		manager:       manager,
		clients:       make(map[string]chan []byte),
		DefaultDomain: "core",
		transactions:  make(map[string]originalIDMapping),
		subscriptions: make(map[string]map[string]bool),
		remoteDomains: make(map[string]*RemoteDomain),
	}
}

// StartRemoteDomain connects to a remote SSE server, discovers its tools, and handles notifications
func (p *Proxy) StartRemoteDomain(name string, sseURL string) {
	rd := &RemoteDomain{
		Name:   name,
		SSEURL: sseURL,
	}
	p.remoteMu.Lock()
	p.remoteDomains[name] = rd
	p.remoteMu.Unlock()

	go func() {
		for {
			log.Printf("[Remote Domain] Connecting to %s at %s\n", name, sseURL)
			req, err := http.NewRequest("GET", sseURL, nil)
			if err != nil {
				log.Printf("[Remote Domain] Error creating request for %s: %v\n", name, err)
				time.Sleep(5 * time.Second)
				continue
			}
			req.Header.Set("Accept", "text/event-stream")

			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				log.Printf("[Remote Domain] Connection failed for %s: %v. Retrying in 5s...\n", name, err)
				time.Sleep(5 * time.Second)
				continue
			}

			scanner := bufio.NewScanner(resp.Body)
			var currentEvent string

			for scanner.Scan() {
				line := scanner.Text()
				if line == "" {
					continue
				}
				if strings.HasPrefix(line, "event:") {
					currentEvent = strings.TrimSpace(strings.TrimPrefix(line, "event:"))
				} else if strings.HasPrefix(line, "data:") {
					dataStr := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
					dataBytes := []byte(dataStr)

					if currentEvent == "endpoint" {
						// Resolve POST endpoint URL
						postPath := dataStr
						if strings.HasPrefix(postPath, "/") {
							if idx := strings.Index(sseURL, "://"); idx != -1 {
								if endIdx := strings.Index(sseURL[idx+3:], "/"); endIdx != -1 {
									rd.PostURL = sseURL[:idx+3+endIdx] + postPath
								} else {
									rd.PostURL = sseURL + postPath
								}
							} else {
								rd.PostURL = postPath
							}
						} else {
							rd.PostURL = postPath
						}
						log.Printf("[Remote Domain] Registered post endpoint for %s: %s\n", name, rd.PostURL)

						// Trigger tool discovery
						go p.discoverRemoteTools(rd)
					} else if currentEvent == "message" {
						p.handleRemoteMessage(name, dataBytes)
					}
				}
			}

			resp.Body.Close()
			log.Printf("[Remote Domain] Connection closed for %s. Reconnecting in 5s...\n", name)
			time.Sleep(5 * time.Second)
		}
	}()
}

func (p *Proxy) discoverRemoteTools(rd *RemoteDomain) {
	time.Sleep(500 * time.Millisecond)

	reqPayload := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      fmt.Sprintf("discover-%s", rd.Name),
		"method":  "tools/list",
		"params":  map[string]interface{}{},
	}

	body, err := json.Marshal(reqPayload)
	if err != nil {
		return
	}

	resp, err := http.Post(rd.PostURL, "application/json", bytes.NewBuffer(body))
	if err != nil {
		log.Printf("[Remote Domain] Failed to send tool discovery to %s: %v\n", rd.Name, err)
		return
	}
	resp.Body.Close()
}

func (p *Proxy) handleRemoteMessage(domainName string, msg []byte) {
	var rpc map[string]interface{}
	if err := json.Unmarshal(msg, &rpc); err == nil {
		// If it's the tools discovery response
		if idVal, exists := rpc["id"]; exists {
			if idStr, ok := idVal.(string); ok && idStr == fmt.Sprintf("discover-%s", domainName) {
				if p.Router != nil {
					p.Router.RegisterDomainTools(domainName, msg)
					log.Printf("[Remote Domain] Registered discovered tools for remote domain %s\n", domainName)
				}
				return
			}
		}

		method, _ := rpc["method"].(string)

		// Handle server-to-client notifications
		if method == "notifications/resources/updated" {
			if params, ok := rpc["params"].(map[string]interface{}); ok {
				if uri, _ := params["uri"].(string); uri != "" {
					p.subMu.Lock()
					sessions := p.subscriptions[uri]
					p.subMu.Unlock()

					p.mu.Lock()
					for sID := range sessions {
						if ch, exists := p.clients[sID]; exists {
							select {
							case ch <- bytes.Clone(msg):
							default:
								log.Printf("[Proxy] Dropped notification for client %s, channel full", sID)
							}
						}
					}
					p.mu.Unlock()
					return // Handled
				}
			}
		}

		// Route responses back to original client
		if idVal, exists := rpc["id"]; exists {
			if txID, ok := idVal.(string); ok && strings.HasPrefix(txID, "tx-") {
				p.txMu.Lock()
				mapping, found := p.transactions[txID]
				if found {
					delete(p.transactions, txID)
				}
				p.txMu.Unlock()

				if found {
					rpc["id"] = mapping.OriginalID
					if rewrittenMsg, err := json.Marshal(rpc); err == nil {
						msg = rewrittenMsg
					}

					p.mu.Lock()
					if ch, exists := p.clients[mapping.SessionID]; exists {
						select {
						case ch <- bytes.Clone(msg):
						default:
							log.Printf("[Proxy] Dropped message for client %s, channel full", mapping.SessionID)
						}
					}
					p.mu.Unlock()
					return // Handled
				}
			}
		}
	}

	// Broadcast logging or progress events
	p.mu.Lock()
	for _, ch := range p.clients {
		select {
		case ch <- bytes.Clone(msg):
		default:
		}
	}
	p.mu.Unlock()
}

// HandleSSE establishes a Server-Sent Events connection for a client
func (p *Proxy) HandleSSE(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	// Generate a session ID
	sessionID := fmt.Sprintf("%d", time.Now().UnixNano())
	messageChan := make(chan []byte, 10)

	p.mu.Lock()
	p.clients[sessionID] = messageChan
	p.mu.Unlock()

	defer func() {
		p.mu.Lock()
		delete(p.clients, sessionID)
		p.mu.Unlock()
		close(messageChan)

		// Cleanup subscriptions for this session
		p.subMu.Lock()
		for uri, sessions := range p.subscriptions {
			if _, exists := sessions[sessionID]; exists {
				delete(sessions, sessionID)
				if len(sessions) == 0 {
					delete(p.subscriptions, uri)
				}
			}
		}
		p.subMu.Unlock()
	}()

	// Send initial endpoint event (MCP SSE spec)
	fmt.Fprintf(w, "event: endpoint\ndata: /message?session_id=%s\n\n", sessionID)
	flusher.Flush()

	for {
		select {
		case msg := <-messageChan:
			fmt.Fprintf(w, "event: message\ndata: %s\n\n", string(msg))
			flusher.Flush()
		case <-r.Context().Done():
			return
		}
	}
}

// HandleMessage receives JSON-RPC requests via POST and routes to a domain process
func (p *Proxy) HandleMessage(w http.ResponseWriter, r *http.Request) {
	sessionID := r.URL.Query().Get("session_id")
	if sessionID == "" {
		http.Error(w, "Missing session_id", http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request", http.StatusInternalServerError)
		return
	}

	// Route based on headers or analyze payload
	domain := r.Header.Get("Mcp-Domain")
	if domain == "" {
		if p.Router != nil {
			if analyzedDomain, err := p.Router.AnalyzeRPC(body); err == nil && analyzedDomain != "" {
				domain = analyzedDomain
			}
		}
		if domain == "" {
			domain = p.DefaultDomain
		}
	}

	// Session isolation rewrite logic
	var rpc map[string]interface{}
	if err := json.Unmarshal(body, &rpc); err == nil {
		method, _ := rpc["method"].(string)

		// Track resource subscriptions
		if method == "resources/subscribe" {
			if params, ok := rpc["params"].(map[string]interface{}); ok {
				if uri, _ := params["uri"].(string); uri != "" {
					p.subMu.Lock()
					if p.subscriptions[uri] == nil {
						p.subscriptions[uri] = make(map[string]bool)
					}
					p.subscriptions[uri][sessionID] = true
					p.subMu.Unlock()
					log.Printf("[Proxy] Session %s subscribed to resource %s", sessionID, uri)
				}
			}
		} else if method == "resources/unsubscribe" {
			if params, ok := rpc["params"].(map[string]interface{}); ok {
				if uri, _ := params["uri"].(string); uri != "" {
					p.subMu.Lock()
					if p.subscriptions[uri] != nil {
						delete(p.subscriptions[uri], sessionID)
					}
					p.subMu.Unlock()
					log.Printf("[Proxy] Session %s unsubscribed from resource %s", sessionID, uri)
				}
			}
		}

		// Rewrite JSON-RPC Request ID if present
		if idVal, exists := rpc["id"]; exists {
			uniqueTxID := fmt.Sprintf("tx-%s-%d", sessionID, time.Now().UnixNano())

			p.txMu.Lock()
			p.transactions[uniqueTxID] = originalIDMapping{
				SessionID:  sessionID,
				OriginalID: idVal,
			}
			p.txMu.Unlock()

			rpc["id"] = uniqueTxID
			if newBody, err := json.Marshal(rpc); err == nil {
				body = newBody
			}
		}
	}

	// CHECK IF REMOTE DOMAIN
	p.remoteMu.RLock()
	rd, isRemote := p.remoteDomains[domain]
	p.remoteMu.RUnlock()

	if isRemote {
		go func() {
			resp, err := http.Post(rd.PostURL, "application/json", bytes.NewBuffer(body))
			if err != nil {
				log.Printf("[Proxy] Error forwarding message to remote domain %s: %v\n", domain, err)
				return
			}
			resp.Body.Close()
		}()
		w.WriteHeader(http.StatusAccepted)
		return
	}

	dp, err := p.manager.GetProcess(domain)
	if err != nil {
		http.Error(w, fmt.Sprintf("Domain %s unavailable: %v", domain, err), http.StatusServiceUnavailable)
		return
	}

	// Write to domain's stdin (adding newline for stdio transport framing)
	dp.Stdin.Write(append(body, '\n'))

	w.WriteHeader(http.StatusAccepted)
}

// StartStdioReader continuously reads JSON-RPC responses from a domain process and routes to clients
func (p *Proxy) StartStdioReader(dp *process.DomainProcess) {
	go func() {
		scanner := bufio.NewScanner(dp.Stdout)
		// Max token size for large JSON payloads
		buf := make([]byte, 0, 64*1024)
		scanner.Buffer(buf, 10*1024*1024)

		for scanner.Scan() {
			msg := scanner.Bytes()

			var rpc map[string]interface{}
			if err := json.Unmarshal(msg, &rpc); err == nil {
				method, _ := rpc["method"].(string)

				// Handle server-to-client notifications
				if method == "notifications/resources/updated" {
					if params, ok := rpc["params"].(map[string]interface{}); ok {
						if uri, _ := params["uri"].(string); uri != "" {
							p.subMu.Lock()
							sessions := p.subscriptions[uri]
							p.subMu.Unlock()

							p.mu.Lock()
							for sID := range sessions {
								if ch, exists := p.clients[sID]; exists {
									select {
									case ch <- bytes.Clone(msg):
									default:
										log.Printf("[Proxy] Dropped notification for client %s, channel full", sID)
									}
								}
							}
							p.mu.Unlock()
							continue // Handled
						}
					}
				}

				// Handle responses (which have "id")
				if idVal, exists := rpc["id"]; exists {
					if txID, ok := idVal.(string); ok && strings.HasPrefix(txID, "tx-") {
						p.txMu.Lock()
						mapping, found := p.transactions[txID]
						if found {
							delete(p.transactions, txID) // cleanup
						}
						p.txMu.Unlock()

						if found {
							// Rewrite back to original client request ID
							rpc["id"] = mapping.OriginalID
							if rewrittenMsg, err := json.Marshal(rpc); err == nil {
								msg = rewrittenMsg
							}

							p.mu.Lock()
							if ch, exists := p.clients[mapping.SessionID]; exists {
								select {
								case ch <- bytes.Clone(msg):
								default:
									log.Printf("[Proxy] Dropped message for client %s, channel full", mapping.SessionID)
								}
							}
							p.mu.Unlock()
							continue // Handled, do not broadcast
						}
					}
				}
			}

			// Broadcast to all connected clients (For logs/progress notifications or fallback)
			p.mu.Lock()
			for _, ch := range p.clients {
				// Non-blocking send
				select {
				case ch <- bytes.Clone(msg):
				default:
					log.Printf("[Proxy] Dropped message for client, channel full")
				}
			}
			p.mu.Unlock()
		}
		if err := scanner.Err(); err != nil {
			log.Printf("[Proxy] Stdio reader error for %s: %v", dp.Name, err)
		}
	}()
}
