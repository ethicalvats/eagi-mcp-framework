---
title: "Core Concepts"
description: "Understand the core architecture, primitives, and modular pattern of the EAGI framework."
---

EAGI divides your application into self-contained, modular folders called **Domains**. Each domain houses its own business services, tools, data resources, prompts, and lifecycle hooks. 

---

## Primitives

The framework wraps standard MCP primitives into type-safe definitions using the TypeScript SDK:

### 1. Tools (Mutations)
Tools represent **actions** the client or AI agent can execute. They accept input validated via Zod schemas and perform operations (such as calling a database or external API).

```typescript
import { defineTool } from '@eagi/sdk';
import { z } from 'zod';

export default defineTool({
  name: 'send_email',
  description: 'Sends a workspace notification email.',
  input: z.object({
    to: z.string().email(),
    body: z.string()
  }),
  handler: async (input, context) => {
    // handler logic
    return { content: [{ type: 'text', text: 'Email sent!' }] };
  }
});
```

### 2. Resources (Queries)
Resources represent **data sources** exposed to the LLM agent or frontend UI. They are queried using URIs (e.g. `tasks://active`).
EAGI supports both **Static Resources** and **Dynamic Resource Templates**:

*   **Static Resource**: A fixed URI (e.g. `billing://terms`) returned in the static resource list.
*   **Resource Template**: A parameterized URI pattern (e.g. `tasks://{id}`). When the client requests `tasks://123`, EAGI's regex matcher parses `{id: "123"}` and injects it into the handler parameters.

```typescript
import { defineResource } from '@eagi/sdk';

export default defineResource({
  uri: 'tasks://{id}',
  name: 'Task details',
  mimeType: 'application/json',
  handler: async (params, context) => {
    return JSON.stringify({ id: params.id, title: `Task number ${params.id}` });
  }
});
```

### 3. Prompts (Templates)
Prompts are pre-defined message templates that provide context or guidelines to LLMs. They can take parameters to format prompts dynamically.

---

## Modular Services & Dependency Injection

To avoid coupling business logic directly to tool handlers, EAGI provides a **topologically sorted service manager**:
*   Define a service in `domains/[domain]/services/`.
*   Declare dependencies on other services.
*   The EAGI Runner initializes all services in dependency order and injects them into the execution context (`context.services`).

```typescript
import { defineService } from '@eagi/sdk';

export default defineService({
  name: 'database',
  deps: ['configService'],
  factory: async (config, deps) => {
    const db = await connectDb(deps.configService.get('DATABASE_URL'));
    return db;
  }
});
```

---

## Hook & Filter Engine

EAGI features a powerful **event hook system** that lets you intercept operations and apply cross-cutting logic:
*   **Actions (`doAction`)**: Event listeners that trigger on lifecycle events (e.g. `before:tool:call`, `after:resource:read`, `on:server:start`). Excellent for auditing and metrics.
*   **Filters (`applyFilters`)**: Middleware that intercepts and transforms data streams (e.g. `filter:tool:input` to redact PII, or `filter:tools:list` to prune schemas).
