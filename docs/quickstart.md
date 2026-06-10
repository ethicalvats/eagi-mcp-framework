---
title: "Quickstart"
description: "Get up and running with the EAGI framework in less than 5 minutes."
---

This guide walks you through setting up an EAGI workspace, defining your first domain server, and running the Go Gateway to expose your tools to clients.

## Prerequisites

Before starting, ensure you have the following installed on your machine:
*   [Node.js](https://nodejs.org/) (v18 or higher)
*   [Go](https://go.dev/) (v1.21 or higher)
*   [pnpm](https://pnpm.io/) (Recommended package manager)

---

## 1. Setup Your Workspace

Initialize a new Node.js project and install the EAGI SDK and CLI packages:

```bash
mkdir my-eagi-workspace
cd my-eagi-workspace
pnpm init
pnpm add @eagi/sdk @eagi/cli zod
```

---

## 2. Define the Config

Create an `eagi.config.ts` in the root of your workspace to configure your gateway and server settings:

```typescript
import { defineConfig } from '@eagi/sdk';

export default defineConfig({
  name: 'my-eagi-workspace',
  version: '0.1.0',
  gateway: {
    port: 3000
  }
});
```

---

## 3. Create a Domain Server

EAGI organizes code into isolated **Domains**. Let's create a `math` domain with a simple add tool.

Create the folder structure:
```bash
mkdir -p domains/math/tools
```

Create a manifest file `domains/math/domain.yaml` to declare your domain:
```yaml
name: math
version: 1.0.0
description: Core math functions for computation.
```

Define the tool in `domains/math/tools/add.ts`:
```typescript
import { defineTool } from '@eagi/sdk';
import { z } from 'zod';

export default defineTool({
  name: 'add',
  description: 'Adds two numbers together.',
  input: z.object({
    a: z.number().describe('The first number'),
    b: z.number().describe('The second number')
  }),
  handler: async (input) => {
    const { a, b } = input;
    return {
      content: [
        {
          type: 'text',
          text: `The sum of ${a} and ${b} is ${a + b}.`
        }
      ]
    };
  }
});
```

---

## 4. Launch the Gateway

Start the Go-based Gateway Control Plane. The CLI will automatically locate your compiled Go gateway binary (or compile it dynamically) and boot up your domain servers:

```bash
npx eagi serve
```

You should see output similar to:
```
Starting EAGI Gateway Control Plane...
[Gateway] Successfully started domain process: math
[Gateway] Registered tools for domain math: [add]
Gateway listening on :3000
```

Congratulations! Your MCP tools are now exposed at `http://localhost:3000/sse` via Server-Sent Events (SSE). You can connect clients like Cursor, Claude, or custom web interfaces directly to this endpoint.
