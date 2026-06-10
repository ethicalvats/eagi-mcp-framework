---
title: "Hooks & Middleware"
description: "Reference guide for customizing execution logic using the EAGI Hook & Filter Engine."
---

The Hook Engine provides a powerful mechanism to intercept, log, and filter data payloads at runtime. Hooks can be registered globally in `eagi.config.ts` or scoped to a specific domain in `domains/[domain]/hooks/index.ts`.

---

## Action Hooks (`doAction`)

Action hooks are event-driven hooks that trigger at specific points in the execution lifecycle. They are run sequentially and cannot alter the data payload.

### Available Actions

| Hook Name | Context Payload | Description |
| :--- | :--- | :--- |
| `before:tool:call` | `{ toolName, input, identity, domain }` | Fires immediately before calling a tool. |
| `after:tool:call` | `{ toolName, input, output, identity, duration }` | Fires immediately after a tool call completes. |
| `on:tool:error` | `{ toolName, input, error, identity }` | Fires if a tool call throws an exception. |
| `before:resource:read`| `{ uri, identity }` | Fires before reading a resource. |
| `after:resource:read` | `{ uri, data, identity }` | Fires after a resource read completes. |
| `on:server:start` | `{ config, domains }` | Fires when the SDK server transport starts. |
| `on:domain:load` | `{ domain }` | Fires when a domain is loaded. |
| `resource:subscribe` | `{ uri }` | Fires when a client subscribes to a resource. |
| `resource:unsubscribe`| `{ uri }` | Fires when a client unsubscribes. |

---

## Filter Hooks (`applyFilters`)

Filter hooks are sequential middleware that can intercept and modify data payloads (like inputs, outputs, or schemas).

### Available Filters

| Hook Name | Input Payload | Output Payload | Description |
| :--- | :--- | :--- | :--- |
| `filter:tool:input` | `unknown` (arguments) | `unknown` (modified arguments) | Modify/redact tool inputs. |
| `filter:tool:output` | `ToolResult` | `ToolResult` (modified output) | Post-process tool results. |
| `filter:resource:data`| `string` | `string` (modified string) | Post-process resource content. |
| `filter:tools:list` | `ToolDefinition[]` | `ToolDefinition[]` (pruned list) | Intercept and prune tools. |

---

## Example Usage

### 1. Registering Scoped Hooks (`domains/math/hooks/index.ts`)

```typescript
export default {
  // Action Hook
  'before:tool:call': async (ctx) => {
    console.log(`[Math Domain] User ${ctx.identity.userId} is calling tool ${ctx.toolName}`);
  },

  // Filter Hook
  'filter:tool:input': async (input, ctx) => {
    // Redact credit card numbers
    if (typeof input === 'object' && input !== null) {
      const copy = { ...input };
      // clean copy...
      return copy;
    }
    return input;
  }
};
```
