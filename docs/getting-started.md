# Getting Started with Enterprise AGI (EAGI)

Welcome to the Enterprise AGI Framework! EAGI is a unified, enterprise-grade framework for building and orchestrating Model Context Protocol (MCP) servers. 

Unlike standard MCP servers, EAGI provides a **Go Control Plane (Gateway)** and a **TypeScript Domain SDK**. This separation gives you the best of both worlds: high-performance multiplexing, routing, and background triggers in Go, combined with the incredibly fast developer experience and ecosystem of TypeScript.

## Quick Start

The best way to understand EAGI is by exploring a working example. We've included a fully rebuilt MCP server called **GapBase** in the `examples/gapbase` directory.

### 1. The Domain Server (TypeScript)

Navigate to the `examples/gapbase` directory. Notice the structure:
- `eagi.config.ts` — Registers the domain configuration.
- `domains/gapbase/services/database.ts` — A shared singleton service that manages your state or database connection.
- `domains/gapbase/tools/` — Each file here exports a strongly-typed tool using `defineTool`.

Tools automatically inherit their `inputSchema` from Zod, making it impossible to write invalid JSON schemas for the LLM.

### 2. Testing Locally via Stdio (Dev Mode)

During development, you can run your Domain Server directly, bypassing the Go Gateway. This exposes standard MCP over `stdio`, meaning it works instantly with Claude Desktop, Claude Code, or Cursor.

\`\`\`bash
cd examples/gapbase
pnpm install
npx eagi serve
\`\`\`

*(See `examples/gapbase/mcp.json` for sample Claude Desktop configurations).*

### 3. Deploying the Control Plane (Production Mesh)

In production, you want Identity Projection (filtering tools by role), Audit Logging, Rate Limiting, and Autonomous background triggers. That's where the Go Gateway comes in.

To boot the Go Gateway and have it orchestrate your GapBase domain:

\`\`\`bash
cd gateway
DOMAIN_DIR=../examples/gapbase go run ./cmd/eagi-gateway
\`\`\`

The Gateway will now:
1. Spawn your TypeScript domain server as a child process.
2. Read the available tools and configure the routing mesh.
3. Expose a unified HTTP/SSE server on `http://localhost:3000`.

You can now point your LLM at `http://localhost:3000/sse`. The Gateway will intercept the requests, enforce authorization, and proxy the JSON-RPC to the appropriate Domain Server.

## Next Steps

- Explore `packages/sdk` to see the underlying TypeScript primitive helpers.
- Explore `gateway/internal` to see how the Go Control Plane orchestrates processes and handles triggers.
- Create your own domain by running `npx @eagi/cli init my-project`!
