package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/eagi/gateway/internal/audit"
	"github.com/eagi/gateway/internal/identity"
	"github.com/eagi/gateway/internal/process"
	"github.com/eagi/gateway/internal/proxy"
	"github.com/eagi/gateway/internal/ratelimit"
	"github.com/eagi/gateway/internal/router"
	"github.com/eagi/gateway/internal/triggers"
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
	triggerEngine := triggers.NewEngine(manager, rtr)

	// 3. Load Domains
	domainDir := os.Getenv("DOMAIN_DIR")
	if domainDir == "" {
		domainDir, _ = os.Getwd()
	}

	if err := manager.StartDomain("core", domainDir); err != nil {
		log.Printf("Warning: Failed to start 'core' domain process: %v\n", err)
	} else {
		// Start stdio reader to pipe MCP output back to the mesh
		if dp, err := manager.GetProcess("core"); err == nil {
			proxyMesh.StartStdioReader(dp)
			// Mocking a registration
			rtr.RegisterDomainTools("core", []byte("{\"result\":{\"tools\":[{\"name\":\"search_gaps\"}]}}"))
		}
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
