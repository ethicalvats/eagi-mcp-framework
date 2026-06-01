package triggers

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/eagi/gateway/internal/process"
	"github.com/eagi/gateway/internal/router"
	"github.com/robfig/cron/v3"
)

// Trigger represents a scheduled autonomous workflow
type Trigger struct {
	Name     string
	Schedule string
	Tool     string
	Input    map[string]interface{}
}

// Engine manages background autonomous executions
type Engine struct {
	cron    *cron.Cron
	manager *process.Manager
	router  *router.Router
}

func NewEngine(manager *process.Manager, router *router.Router) *Engine {
	return &Engine{
		cron:    cron.New(cron.WithSeconds()), // Standard cron with seconds precision
		manager: manager,
		router:  router,
	}
}

// RegisterTrigger schedules a new cron job for a tool execution
func (e *Engine) RegisterTrigger(t Trigger) error {
	_, err := e.cron.AddFunc(t.Schedule, func() {
		e.executeTrigger(t)
	})
	if err != nil {
		return fmt.Errorf("failed to schedule trigger %s: %v", t.Name, err)
	}

	log.Printf("[Triggers] Registered autonomous trigger '%s' on schedule: %s\n", t.Name, t.Schedule)
	return nil
}

func (e *Engine) Start() {
	e.cron.Start()
	log.Println("[Triggers] Cron scheduler started")
}

func (e *Engine) Stop() {
	e.cron.Stop()
	log.Println("[Triggers] Cron scheduler stopped")
}

func (e *Engine) executeTrigger(t Trigger) {
	log.Printf("[Triggers] Executing '%s' -> calling tool %s\n", t.Name, t.Tool)

	// Determine which domain handles this tool
	domainName, err := e.router.GetDomainForTool(t.Tool)
	if err != nil {
		log.Printf("[Triggers] Error executing trigger %s: %v\n", t.Name, err)
		return
	}

	dp, err := e.manager.GetProcess(domainName)
	if err != nil {
		log.Printf("[Triggers] Error getting domain process for trigger %s: %v\n", t.Name, err)
		return
	}

	// Craft a JSON-RPC tools/call request
	req := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      fmt.Sprintf("trigger-%s", t.Name),
		"method":  "tools/call",
		"params": map[string]interface{}{
			"name":      t.Tool,
			"arguments": t.Input,
		},
	}

	payload, err := json.Marshal(req)
	if err != nil {
		log.Printf("[Triggers] Error marshaling request for trigger %s: %v\n", t.Name, err)
		return
	}

	// Write to domain's stdin
	if _, err := dp.Stdin.Write(append(payload, '\n')); err != nil {
		log.Printf("[Triggers] Error writing to domain process for trigger %s: %v\n", t.Name, err)
	}
}
