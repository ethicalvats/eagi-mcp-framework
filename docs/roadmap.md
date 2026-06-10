---
title: "Roadmap"
description: "The technical roadmap to bring the EAGI framework to full feature parity with enterprise MCP standards."
---

The development of the EAGI framework is structured in three phases to systematically address security, spec compliance, and enterprise operational capabilities.

---

## Phase 1: High Priority (Completed)
*   **Response Routing & Session Isolation**: Rewrote JSON-RPC IDs to unique gateway transaction IDs to prevent client response leakage.
*   **Dynamic Resource Templates**: Added `ListResourceTemplatesRequestSchema` and parameter parsing/regex matching.
*   **SSE Subscriptions**: Implemented stateful `Subscribe` and `Unsubscribe` routing for real-time resource updates.
*   **Remote SSE Server Routing**: Allowed proxying requests to external HTTP/SSE MCP servers via `gateway.config.json`.

---

## Phase 2: Medium Priority (In Progress)
*   **Asynchronous Tasks (`io.modelcontextprotocol/tasks`)**: Standard compliance for long-running workflows with terminal statuses (`completed`, `failed`, `cancelled`), task-specific cancellation, and metadata mapping.
*   **OAuth 2.1 & Client ID Metadata Documents (CIMD)**: Shift from pre-shared HMAC keys to stateless token verification and remote metadata discovery via CIMD client identifier URLs.
*   **Tool Groups / Virtual Servers**: Configuration-driven capabilities mapping roles to specific subsets of tools, prompts, and resources.

---

## Phase 3: Low Priority (Advanced Features)
*   **Dynamic Tool Retrieval (Tool-RAG)**: Built-in vector search filtering that semantically selects the most relevant tools for the LLM context query.
*   **Enterprise Database Backend**: Transition from in-memory routing states to PostgreSQL to support multi-tenant sessions and persistent task statuses.
*   **OpenTelemetry Observability**: Centralized tracing and performance metrics for tools and resources.
*   **Tamper-Proof Audit Chaining**: Cryptographically chained hashing of audit logs to guarantee compliance records cannot be modified.
