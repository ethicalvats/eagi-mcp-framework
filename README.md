# EAGI (eagi-mcp-framework): Enterprise Model Context Protocol Framework

EAGI (Enterprise AGI) is a unified, production-grade framework for building and orchestrating Model Context Protocol (MCP) servers. 

Designed for scalability, auditability, and safety, EAGI splits responsibilities between a **Go Control Plane (Gateway)** and a **TypeScript Domain SDK**. This hybrid architecture offers high-performance multiplexing, routing, and background scheduling in Go alongside a fast, type-safe developer experience in TypeScript.

---

## Key Features

- 🔌 **WordPress-style Hooks**: Robust hook engine supporting synchronous and asynchronous actions and filters (`before:tool:call`, `filter:tool:output`, etc.) to extend domain logic without modifying core server code.
- 🛡️ **Identity Projection**: Enforce Role-Based Access Control (RBAC) at the Gateway. Authenticate users via OAuth/JWT, map identities to roles, and filter available tools dynamically before the LLM sees them.
- 🗃️ **Zod-to-MCP Schemas**: Define tool inputs using standard Zod schemas; the SDK automatically builds and validates compliant JSON schemas.
- ⏱️ **Autonomous Triggers**: High-performance cron scheduler built into the Go gateway to run background agent workflows autonomously.
- 📝 **Compliance-grade Audit Logging**: Built-in cryptographic hash-chained audit logging and field-redaction middleware.
- ⚡ **Go process manager & proxy**: Spawns long-lived domain Node.js processes, handles communications over `stdio` transport, and exposes them as a unified `HTTP/SSE` mesh.

---

## Repository Structure

```
eagi/
├── package.json          # Monorepo configuration (pnpm workspaces)
├── packages/
│   ├── sdk/              # @eagi/sdk — Core framework SDK (TypeScript)
│   ├── cli/              # @eagi/cli — CLI tools (`eagi serve`, `eagi dev`)
│   └── create-eagi/      # create-eagi — Scaffolding bootstrapper
├── gateway/              # Go-based Control Plane/Gateway
├── examples/
│   └── gapbase/          # Sample domain server for testing
└── docs/                 # Getting started and primitives guides
```

---

## Quick Start

### 1. Requirements

Make sure you have installed:
- [Node.js](https://nodejs.org/) (v18+)
- [Go](https://go.dev/) (v1.20+)
- [pnpm](https://pnpm.io/)

### 2. Installation

Clone this repository and install the dependencies:

```bash
pnpm install
pnpm build
```

### 3. Running GapBase Example

To test the monorepo with the included `gapbase` example:

#### Developer Mode (Stdio Transport)

Run the Domain SDK directly over standard input/output. This is fully compatible with local LLM environments like Claude Desktop or Claude Code:

```bash
cd examples/gapbase
npx eagi serve
```

#### Production Mode (Go Gateway + SSE Transport)

Run the Go Control Plane to launch and orchestrate the domain. This starts the SSE server proxying requests to the background domain processes:

```bash
cd gateway
DOMAIN_DIR=../examples/gapbase go run ./cmd/eagi-gateway
```

The gateway exposes a unified MCP SSE endpoint at `http://localhost:3000/sse` with identity mapping, triggers, and rate-limiting enabled.

---

## Documentation

For deep dives into EAGI architecture and building custom domains:
- 📖 [Getting Started Guide](./docs/getting-started.md)
- 🔑 [Advanced Hooks & Identity Guide](./docs/hooks-and-identity.md)
- 🌐 [Gateway Reference](./docs/gateway.md) *(refer to gateway codebase)*

---

## Enterprise MCP Framework Comparison

EAGI is designed specifically for enterprise-grade deployments. Below is a comparison of EAGI against other major SDKs and frameworks in the MCP ecosystem:

| Feature | EAGI (eagi-mcp-framework) | Spring AI MCP (Java) | McpDotNet (C#) | Official SDKs (JS/Python) | FastMCP (Python) |
| :--- | :--- | :--- | :--- | :--- | :--- |
| **Primary Language** | Go (Gateway) + TS (SDK) | Java | C# / .NET | TypeScript / Python | Python |
| **Orchestration Mesh** | **Yes** (spawns & monitors child processes) | No (client only) | No (client only) | No (single server) | No (single server) |
| **Extension Hooks** | **Yes** (Actions & Filters engine) | No | No | No | No |
| **Identity / RBAC** | **Yes** (projects OAuth JWT to tool filters) | No (relies on Spring Security) | No | No | No |
| **Background Triggers** | **Yes** (built-in cron / webhook runners) | No | No | No | No |
| **Observability** | **Yes** (structured hash-chained audit logs) | Basic logging | Basic logging | Basic logging | Basic logging |
| **Target Persona** | Production Platforms | Java Enterprise Apps | C#/.NET Enterprise Apps | Protocol Developers | Rapid Prototyping |

---

## License

This project is licensed under the Sustainable Use License (Fair-Code). See [LICENSE](./LICENSE) for details. For commercial hosting, SaaS integration, or white-labeled redistribution, please contact us for a commercial license.
