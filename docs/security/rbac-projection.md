---
title: "Identity Projection & RBAC"
description: "Dynamically prune exposed tools and authorize actions based on user roles and OIDC claims."
---

AI models perform better when they are presented with a concise, relevant list of tools rather than hundreds of options (avoiding context bloat). EAGI handles this by using **Identity Projection** inside the Go Gateway to prune the tool list before the LLM ever sees it.

---

## Token Authentication

The Go Gateway intercepts OIDC / JWT Bearer tokens from incoming client request headers:

1.  **JWT Verification**: The `ProjectionEngine` validates the signature and expiration of the JWT.
2.  **Claim Propagation**: Extracts roles and user claims, and injects them as downstream headers:
    *   `X-EAGI-User`: The authenticated user ID (`sub`).
    *   `X-EAGI-Role`: The user's primary authorization role (e.g. `admin`, `viewer`).
    *   `X-EAGI-Claims`: JSON string containing OIDC claims.
3.  **Context Injection**: The TypeScript domain SDK parses these headers and populates `context.identity` for tool and resource handlers.

---

## Tool List Projection (RBAC Pruning)

When a client queries the gateway for available tools (`tools/list`), the gateway evaluates the tools list against the user's role:

*   **Policy Definition**: Domain manifests (`domain.yaml`) can specify authorization policies.
*   **Dynamic Pruning**: If a user has a `viewer` role, the gateway intercepts the JSON-RPC response and strips out administrative or destructive tools (e.g. `delete_user` or `approve_invoice`).
*   **Benefits**:
    *   **Reduced Context Window**: Less tokens are wasted describing schemas to the LLM.
    *   **Enhanced Reasoning**: The LLM avoids selecting tools it does not have permission to run.
    *   **Proactive Security**: Tools are hidden from unauthorized agents, preventing prompt injection attacks targeting unauthorized capabilities.
