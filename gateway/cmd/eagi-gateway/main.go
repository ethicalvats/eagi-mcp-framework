package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/eagi/gateway/internal/audit"
	"github.com/eagi/gateway/internal/identity"
	"github.com/eagi/gateway/internal/process"
	"github.com/eagi/gateway/internal/proxy"
	"github.com/eagi/gateway/internal/ratelimit"
	"github.com/eagi/gateway/internal/router"
	"github.com/eagi/gateway/internal/triggers"
	"github.com/gorilla/mux"
)

func main() {
	log.Println("Starting EAGI Gateway (Control Plane)...")

	// 1. Initialize core components
	manager := process.NewManager()
	rtr := router.NewRouter()

	// Read JWT secret from env
	secret := os.Getenv("EAGI_JWT_SECRET")
	if secret == "" {
		secret = "dev-secret-do-not-use-in-prod"
	}
	identityEngine := identity.NewProjectionEngine(secret)

	auditLogger, err := audit.NewLogger("stdout")
	if err != nil {
		log.Fatalf("Failed to initialize audit logger: %v", err)
	}

	limiter := ratelimit.NewLimiter()

	// 2. Initialize Proxy Mesh and Triggers
	proxyMesh := proxy.NewProxy(manager)
	proxyMesh.Router = rtr
	triggerEngine := triggers.NewEngine(manager, rtr)

	// 3. Load Domains
	domainDir := os.Getenv("DOMAIN_DIR")
	if domainDir == "" {
		domainDir, _ = os.Getwd()
	}

	// Load gateway.config.json for remote domains
	type GatewayConfig struct {
		RemoteDomains map[string]string `json:"remoteDomains"`
	}
	var gatewayConfig GatewayConfig
	configFile := filepath.Join(domainDir, "gateway.config.json")
	if _, err := os.Stat(configFile); err == nil {
		data, err := os.ReadFile(configFile)
		if err == nil {
			if err := json.Unmarshal(data, &gatewayConfig); err != nil {
				log.Printf("Warning: Failed to parse config file: %v\n", err)
			} else {
				log.Printf("[Gateway] Loaded config: %d remote domains\n", len(gatewayConfig.RemoteDomains))
			}
		}
	}

	// Connect and start remote domains
	for domainName, sseURL := range gatewayConfig.RemoteDomains {
		proxyMesh.StartRemoteDomain(domainName, sseURL)
		log.Printf("[Gateway] Configured remote domain '%s' -> %s\n", domainName, sseURL)
	}

	domainsPath := filepath.Join(domainDir, "domains")
	entries, err := os.ReadDir(domainsPath)
	if err != nil {
		log.Printf("Warning: Failed to read domains directory %s: %v\n", domainsPath, err)
	}

	var bootedDomains []string
	for _, entry := range entries {
		if entry.IsDir() {
			domainName := entry.Name()
			if err := manager.StartDomain(domainName, domainDir); err != nil {
				log.Printf("Warning: Failed to start '%s' domain process: %v\n", domainName, err)
				continue
			}

			bootedDomains = append(bootedDomains, domainName)
			log.Printf("[Gateway] Successfully started domain process: %s\n", domainName)

			dp, err := manager.GetProcess(domainName)
			if err != nil {
				log.Printf("Warning: Failed to retrieve process for %s: %v\n", domainName, err)
				continue
			}

			proxyMesh.StartStdioReader(dp)

			// Discover tools dynamically by reading tools directory
			toolsPath := filepath.Join(domainsPath, domainName, "tools")
			toolEntries, err := os.ReadDir(toolsPath)
			var toolNames []string
			if err == nil {
				for _, toolEntry := range toolEntries {
					if !toolEntry.IsDir() {
						filename := toolEntry.Name()
						// Check if TS or JS file
						if len(filename) > 3 && (filename[len(filename)-3:] == ".ts" || filename[len(filename)-3:] == ".js") {
							toolName := filename[:len(filename)-3]
							toolNames = append(toolNames, toolName)
						}
					}
				}
			}

			// Register discovered tools
			if len(toolNames) > 0 {
				type ToolItem struct {
					Name string `json:"name"`
				}
				type ToolsResult struct {
					Tools []ToolItem `json:"tools"`
				}
				type ToolsResponse struct {
					Result ToolsResult `json:"result"`
				}

				respObj := ToolsResponse{
					Result: ToolsResult{
						Tools: make([]ToolItem, len(toolNames)),
					},
				}
				for i, tName := range toolNames {
					respObj.Result.Tools[i] = ToolItem{Name: tName}
				}

				rawToolsList, err := json.Marshal(respObj)
				if err == nil {
					rtr.RegisterDomainTools(domainName, rawToolsList)
					log.Printf("[Gateway] Registered tools for domain %s: %v\n", domainName, toolNames)
				}
			}
		}
	}

	if len(bootedDomains) > 0 {
		rtr.DefaultDomain = bootedDomains[0]
		proxyMesh.DefaultDomain = bootedDomains[0]
		log.Printf("[Gateway] Default fallback domain set to: %s\n", bootedDomains[0])
	}

	// Start Trigger Engine
	triggerEngine.Start()
	defer triggerEngine.Stop()

	// 4. Setup HTTP Server
	r := mux.NewRouter()

	// Middleware
	r.Use(identityEngine.Authenticate)
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			userID := req.Header.Get("X-EAGI-User")
			role := req.Header.Get("X-EAGI-Role")
			if err := limiter.CheckLimit(userID, role); err != nil {
				auditLogger.Log(audit.AuditEntry{
					Domain: "gateway",
					UserID: userID,
					Role:   role,
					Status: http.StatusTooManyRequests,
				})
				http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
				return
			}
			next.ServeHTTP(w, req)
		})
	})

	// MCP SSE endpoints
	r.HandleFunc("/sse", proxyMesh.HandleSSE).Methods("GET")
	r.HandleFunc("/message", proxyMesh.HandleMessage).Methods("POST")

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	log.Printf("Gateway listening on :%s\n", port)
	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
