package router

import (
	"encoding/json"
	"fmt"
	"sync"
)

// Router maintains the mapping between tool names and domain processes
type Router struct {
	toolToDomain  map[string]string // e.g. "approve_invoice" -> "invoicing"
	DefaultDomain string
	mu            sync.RWMutex
}

func NewRouter() *Router {
	return &Router{
		toolToDomain:  make(map[string]string),
		DefaultDomain: "core",
	}
}

// RegisterDomainTools updates the routing table with tools from a specific domain
func (r *Router) RegisterDomainTools(domainName string, rawToolsList []byte) error {
	var response map[string]interface{}
	if err := json.Unmarshal(rawToolsList, &response); err != nil {
		return err
	}

	result, ok := response["result"].(map[string]interface{})
	if !ok {
		return nil
	}

	tools, ok := result["tools"].([]interface{})
	if !ok {
		return nil
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	for _, t := range tools {
		toolMap := t.(map[string]interface{})
		toolName, ok := toolMap["name"].(string)
		if ok {
			r.toolToDomain[toolName] = domainName
		}
	}

	return nil
}

// GetDomainForTool returns the domain name responsible for a given tool
func (r *Router) GetDomainForTool(toolName string) (string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	domain, exists := r.toolToDomain[toolName]
	if !exists {
		return "", fmt.Errorf("no domain registered for tool: %s", toolName)
	}

	return domain, nil
}

// AnalyzeRPC payload to determine destination domain
func (r *Router) AnalyzeRPC(payload []byte) (string, error) {
	var rpc map[string]interface{}
	if err := json.Unmarshal(payload, &rpc); err != nil {
		return "", err
	}

	method, _ := rpc["method"].(string)

	if method == "tools/call" {
		params, ok := rpc["params"].(map[string]interface{})
		if !ok {
			return "", fmt.Errorf("missing params in tools/call")
		}
		toolName, _ := params["name"].(string)
		return r.GetDomainForTool(toolName)
	}

	// For resources or prompts, similar routing logic applies.
	// For MVP, if it's not a tool call, we default to the DefaultDomain or reject
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.DefaultDomain, nil
}
