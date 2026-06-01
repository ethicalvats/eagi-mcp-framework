# Contributing to EAGI (eagi-mcp-framework)

Thank you for your interest in contributing to EAGI! We welcome contributions from the community to help make this framework the best enterprise-grade Model Context Protocol (MCP) orchestrator.

---

## Code of Conduct

By participating in this project, you agree to abide by our [Code of Conduct](./CODE_OF_CONDUCT.md).

---

## Getting Started

### Prerequisites

To build and run the framework locally, you need:
- [Node.js](https://nodejs.org/) (v18+)
- [Go](https://go.dev/) (v1.20+)
- [pnpm](https://pnpm.io/) (v9+)

### Local Development Setup

1. Clone the repository and navigate into it:
   ```bash
   git clone https://github.com/username/eagi-mcp-framework.git
   cd eagi-mcp-framework
   ```

2. Install dependencies for the TypeScript monorepo:
   ```bash
   pnpm install
   ```

3. Build all TypeScript packages (SDK, CLI, and bootstrappers):
   ```bash
   pnpm build
   ```

4. Verify tests run successfully:
   - TypeScript unit tests:
     ```bash
     pnpm test
     ```
   - Go gateway tests/compilation:
     ```bash
     cd gateway
     go test ./...
     go build -o bin/eagi-gateway ./cmd/eagi-gateway
     ```

---

## Development Workflow

1. **Find an Issue**: Search our open issues or file a new one to discuss your proposed feature or fix.
2. **Create a Branch**: Fork the repo and create a branch for your changes:
   ```bash
   git checkout -b feature/my-amazing-feature
   ```
3. **Write Tests**: Ensure your feature is covered by tests.
4. **Code Quality**:
   - Format TypeScript code: `pnpm format`
   - Lint TypeScript: `pnpm lint`
   - Run Go format: `go fmt ./...` inside `/gateway`
5. **Submit a Pull Request**: Push your branch to GitHub and open a PR against our `main` branch. Provide a clear description of the problem and your solution.

Thank you for making EAGI better for everyone!
