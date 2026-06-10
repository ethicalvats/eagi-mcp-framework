---
title: "Remote SSE Routing"
description: "Configure the EAGI Go Gateway to route requests to remote MCP servers over HTTP/SSE."
---

While Local Mode is excellent for development, production workloads often require routing requests to remote microservices or external MCP servers. The EAGI Go Gateway supports **Remote Mode** out of the box using configuration-driven HTTP/SSE routing.

---

## 1. Configure Remote Domains

Create a `gateway.config.json` file in your workspace working directory. Register your remote domains and their corresponding SSE endpoints:

```json
{
  "remoteDomains": {
    "billing": "http://billing-service.internal/sse",
    "crm": "http://crm-service.internal/sse"
  }
}
```

When the gateway starts, it reads this file, bypasses local process spawning for these domains, and connects directly as a client to the remote servers.

---

## 2. Dynamic Capabilities Discovery

For every configured remote domain, the Go Gateway performs an **asynchronous handshake**:

1.  **Connection**: Opens a background HTTP GET connection to the remote SSE endpoint (e.g. `http://billing-service.internal/sse`).
2.  **Endpoint Resolution**: Listens for the remote server's `event: endpoint` payload. It resolves the POST endpoint URL (e.g. `http://billing-service.internal/message?session_id=...`).
3.  **Tool Discovery**: Sends an HTTP POST `tools/list` request to the resolved POST endpoint. When the response arrives over the SSE connection, the gateway parses the list of tools and dynamically registers them in the gateway router.

---

## 3. Proxying Execution

When a client calls a remote tool (e.g. `billing/pay_invoice`), the gateway router resolves it to the remote domain:

```mermaid
sequenceDiagram
    participant Client
    participant GW as EAGI Gateway
    participant RS as Remote MCP Server
    
    Client->>GW: POST /message?session_id=A (tool call ID: 1)
    GW->>GW: Rewrite ID 1 to unique tx-A-999
    GW->>RS: POST /message (tool call ID: tx-A-999)
    RS-->>GW: HTTP 202 Accepted
    GW-->>Client: HTTP 202 Accepted
    Note over RS, GW: Remote execution runs...
    RS->>GW: SSE Event: message (Response ID: tx-A-999)
    GW->>GW: Match tx-A-999 -> Client A, restore ID to 1
    GW->>Client: SSE Event: message (Response ID: 1)
```

This ensures complete decoupling: remote servers execute tool calls independently, and the gateway proxies the payloads with low latency.
