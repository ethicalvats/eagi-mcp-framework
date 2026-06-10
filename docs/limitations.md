---
title: "Limitations"
description: "Current design boundaries and limitations of the EAGI framework."
---

While EAGI is built for enterprise-grade workloads, certain features are currently in development or limited by current design scope. 

---

## 1. Asynchronous Tasks Compliance
The 2025-11-25 MCP tasks primitive (`io.modelcontextprotocol/tasks`) is currently **unsupported** in the core. All tool executions are synchronous and block the SSE response stream until completed. Very long tool executions (exceeding 2–5 minutes) can cause gateway timeouts.

---

## 2. In-Memory Routing Table
The Go Gateway maintains the routing maps (tool name to domain, active transactions, and resource subscriptions) entirely in memory.
*   **Non-Persistent**: If the gateway process restarts, active client subscriptions and running transaction mapping records are lost.
*   **Scale Limitation**: High Availability (HA) deployments with multiple gateway instances behind a load balancer are not supported in Local Mode since states are not shared.

---

## 3. Pre-Shared JWT Auth
Current authentication in Go Gateway expects a symmetric pre-shared HMAC secret.
*   **OIDC/JWKS**: There is no native support to fetch public keys from OIDC providers dynamically.
*   **Stateless Registrations**: The gateway does not support Client ID Metadata Documents (CIMD) for on-demand OAuth client registrations.

---

## 4. Single-Tenant Process Manager
The Local Process Manager spawns Node.js processes globally for the server. There is no sandbox or container isolation per client session. All users calling tools in a domain share the same OS execution namespace.
