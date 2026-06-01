package proxy

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/eagi/gateway/internal/process"
)

// Proxy handles HTTP/SSE MCP requests and proxies them to stdio Node processes
type Proxy struct {
	manager  *process.Manager
	clients  map[string]chan []byte // SSE connections
	mu       sync.Mutex
}

func NewProxy(manager *process.Manager) *Proxy {
	return &Proxy{
		manager: manager,
		clients: make(map[string]chan []byte),
	}
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

	// Route based on headers (July 2026 stateless spec)
	// Example: Mcp-Domain: invoicing
	domain := r.Header.Get("Mcp-Domain")
	if domain == "" {
		// Fallback to a default domain or reject
		domain = "core" 
	}

	dp, err := p.manager.GetProcess(domain)
	if err != nil {
		http.Error(w, fmt.Sprintf("Domain %s unavailable: %v", domain, err), http.StatusServiceUnavailable)
		return
	}

	// Write to domain's stdin (adding newline for stdio transport framing)
	dp.Stdin.Write(append(body, '\n'))

	// The process's stdout needs to be continuously read and broadcast to the SSE channel.
	// This is handled by a background reader per domain.
	
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
			
			// Broadcast to all connected clients (In a real setup, we route based on JSON-RPC ID)
			// For this MVP, we broadcast to all active SSE connections.
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
