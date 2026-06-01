# Scaling MCP to 1000+ Tools: Context Optimization & Orchestration

One of the most common bottlenecks when scaling Model Context Protocol (MCP) implementations in enterprise environments is **tool bloat**. As organizations add more services and features, the total number of tools can easily reach 100 or even 1000+. 

Exposing all tool schemas directly to an LLM on every query causes several severe issues:
1. **Context Window Exhaustion**: Hundreds of tool definitions (names, descriptions, and parameter schemas) consume a massive amount of prompt tokens.
2. **High Latency & Costs**: Unnecessary prompt tokens increase inference latency and API costs.
3. **Reasoning Degradation ("Context Rot")**: LLMs lose reasoning accuracy and struggle to select the correct tool when overwhelmed by hundreds of choices.

EAGI natively addresses this scaling challenge through its **hybrid architecture (Go Control Plane + Domain SDK)** using the following features:

---

## 1. Identity Projection (RBAC Tool Pruning)

In an enterprise setting, a single user session or agent workflow rarely needs access to all corporate tools. 

EAGI's **Go Gateway** acts as a stateless, high-performance security guard:
* **Token Interception**: The Gateway intercepts the client's `tools/list` request and parses the user's OAuth/JWT token.
* **Role-Based Pruning**: Using the gateway's **`ProjectionEngine`**, it dynamically filters the list of tools before the client or LLM ever sees them. For example, if a user has the `finance` role, the gateway filters out developer operations and marketing tools.
* **Security & Optimization**: By restricting visibility to authorized scopes, the total tool payload is pruned from 1000+ to just the small subset (e.g., 5–15 tools) relevant to the user's role.

---

## 2. Dynamic Tool Retrieval (Tool-RAG via Hooks)

For scenarios where a single role still has access to hundreds of tools, EAGI supports **dynamic tool retrieval (Tool-RAG)** using its built-in Hook & Filter engine.

Instead of loading all tools, the Gateway can load them dynamically based on query intent:
1. The LLM agent sends the user's latest query message.
2. The Gateway intercepts the request using a `filter:tools:list` filter hook.
3. The hook runs a lightweight semantic search (Vector Search) comparing the user's query against an index of tool descriptions.
4. The hook filters the returned list, leaving only the top 3–5 most semantically relevant tools in the response sent to the LLM.

### Example Hook Concept
You can configure a filter hook in your gateway config or custom extensions:

```typescript
import { registerHook } from '@eagi/sdk';

// Filter the tool list dynamically based on the current context/query
registerHook('filter:tools:list', async (tools, context) => {
  const userQuery = context.session.lastUserMessage;
  if (!userQuery) return tools; // Return all tools if no query is present

  // 1. Fetch semantically relevant tool names from your vector database
  const relevantToolNames = await vectorDb.search(userQuery, {
    index: 'tool-descriptions',
    limit: 5
  });

  // 2. Return only the tools matching the relevant search results
  return tools.filter(tool => relevantToolNames.includes(tool.name));
});
```

---

## 3. Micro-service Domain Partitioning

EAGI separates the **Go Control Plane (Gateway)** from the **TypeScript Domain SDK**. This allows you to split a monolithic codebase into multiple, isolated **Domain Servers** running as lightweight background processes:
* **Independent Execution**: You can create separate Domain Servers for different business units (e.g., `billing-server`, `inventory-server`, `crm-server`).
* **Route Namespaces**: The Gateway can expose distinct namespaces (e.g., `/sse/crm`, `/sse/finance`) rather than merging everything into one monolithic transport. This allows clients to target their context explicitly, further reducing tool-resolution namespaces.

---

## 4. Low-Latency Gateway (Go Control Plane)

Intercepting, parsing, and filtering tool lists on every request adds overhead. Performing these complex operations inside a single-threaded runtime (like Node.js) or a slower runtime under heavy load degrades performance.

EAGI solves this by implementing the orchestration mesh in **Go**:
* The Go Control Plane handles child process management, JWT decoding, RBAC filtering, and hook routing with sub-millisecond overhead.
* The heavier TypeScript Domain Servers are only spun up or invoked when a specific tool handler actually needs to be executed, keeping the system highly performant and responsive.
