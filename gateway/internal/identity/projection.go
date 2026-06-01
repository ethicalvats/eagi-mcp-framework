package identity

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

// Identity represents the authenticated user and their capabilities
type Identity struct {
	UserID string
	Role   string
	Claims map[string]interface{}
}

// ProjectionEngine handles OAuth token verification and capability projection
type ProjectionEngine struct {
	JWTSecret []byte
}

func NewProjectionEngine(secret string) *ProjectionEngine {
	return &ProjectionEngine{
		JWTSecret: []byte(secret),
	}
}

// Authenticate middleware verifies the JWT and injects Identity into the request context
func (p *ProjectionEngine) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			// For this MVP, if no auth, we proceed as 'anonymous' or 'developer'
			r.Header.Set("X-EAGI-Role", "developer")
			r.Header.Set("X-EAGI-User", "anonymous")
			next.ServeHTTP(w, r)
			return
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")

		token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return p.JWTSecret, nil
		})

		if err != nil || !token.Valid {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			userID, _ := claims["sub"].(string)
			role, _ := claims["role"].(string)

			if role == "" {
				role = "viewer" // default
			}

			// Pass identity down to the proxy via headers (stateless)
			r.Header.Set("X-EAGI-Role", role)
			r.Header.Set("X-EAGI-User", userID)

			// Optional: Encode full claims as JSON header
			claimsJSON, _ := json.Marshal(claims)
			r.Header.Set("X-EAGI-Claims", string(claimsJSON))
		}

		next.ServeHTTP(w, r)
	})
}

// ProjectToolsList filters the raw `tools/list` response from the domain server
// based on the user's role and the domain's authorization policy.
// This ensures the LLM never even sees tools the user isn't authorized to use.
func (p *ProjectionEngine) ProjectToolsList(rawJSON []byte, role string) ([]byte, error) {
	// 1. Parse raw JSON-RPC response containing `tools` array
	var response map[string]interface{}
	if err := json.Unmarshal(rawJSON, &response); err != nil {
		return nil, err
	}

	result, ok := response["result"].(map[string]interface{})
	if !ok {
		return rawJSON, nil // Not a standard tools/list response
	}

	tools, ok := result["tools"].([]interface{})
	if !ok {
		return rawJSON, nil
	}

	// 2. Filter tools (In a real implementation, we cross-reference domain.yaml policies)
	// For MVP: We assume the Node domain server handles granular auth on execution,
	// but identity projection is about NOT SHOWING them.
	// Since we don't have the domain.yaml loaded into Go memory for this simple example,
	// we skip deep filtering. In production, Gateway loads `domain.yaml` and does RBAC.

	filteredTools := make([]interface{}, 0)
	for _, t := range tools {
		toolMap := t.(map[string]interface{})
		// Example filter: if role is viewer, hide tools with "approve" or "create"
		toolName := toolMap["name"].(string)
		if role == "viewer" && (strings.Contains(toolName, "approve") || strings.Contains(toolName, "create")) {
			continue // Hide
		}
		filteredTools = append(filteredTools, t)
	}

	result["tools"] = filteredTools
	response["result"] = result

	return json.Marshal(response)
}
