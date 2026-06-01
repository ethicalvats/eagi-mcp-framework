# GapBase MCP Example

This is a demonstration of how to rebuild an existing Node.js MCP server (GapBase) using the **Enterprise AGI (EAGI) Framework**.

## Overview

The original `sample-mcp.js` was a monolithic file with a massive switch statement routing JSON-RPC tool calls.

By rebuilding it in the EAGI framework, we gain:
1. **Separation of Concerns**: Each tool is an isolated, type-safe file using `defineTool`.
2. **Type Safety**: Zod schema definitions automatically map to MCP `inputSchema` and infer the `input` type in the handler.
3. **Dependency Injection**: The `GapBaseService` loads the JSON files once and is injected into all tool handlers via `ctx.services.gapbase`.
4. **Control Plane Readiness**: This domain is ready to be hosted by the Go Gateway, gaining Identity Projection, Audit Logging, and Autonomous Triggers out-of-the-box.

## Project Structure

```text
examples/gapbase/
├── eagi.config.ts                     # Framework configuration
├── package.json
├── data/
│   ├── gaps.json                      # Mock database
│   └── trends.json
└── domains/
    └── gapbase/
        ├── domain.yaml                # Domain manifest
        ├── services/
        │   └── database.ts            # Shared GapBaseService (Singleton dependency)
        └── tools/                     # Isolated, type-safe tool handlers
            ├── get_gap.ts
            ├── get_stats.ts
            ├── get_viral.ts
            ├── list_industries.ts
            └── search_gaps.ts
```

## Running & Testing

1. **Install dependencies**
   ```bash
   pnpm install
   ```

2. **Run the local MCP Server (Node-only testing)**
   You can run the domain server directly using the CLI:
   ```bash
   npm run serve
   ```
   This will boot the EAGI SDK runner and connect the Stdio transport. You can point Claude Code or any MCP client directly to:
   `node node_modules/.bin/eagi serve`

3. **Run via the Go Gateway (Production Mesh)**
   To leverage the full Control Plane:
   ```bash
   # From the repository root
   cd gateway
   go run ./cmd/eagi-gateway
   ```
   The Gateway will spawn the Node process as a child, monitor it, and expose the HTTP/SSE proxy on port 3000.
