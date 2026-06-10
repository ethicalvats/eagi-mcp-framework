---
title: "Define Tools"
description: "Reference guide for creating and configuring MCP Tools using the EAGI SDK."
---

Tools are defined as standard files under `domains/[domain]/tools/` and exported as default imports using the `defineTool` helper function.

---

## Tool Definition Schema

A tool configuration accepts the following properties:

| Property | Type | Required | Description |
| :--- | :--- | :--- | :--- |
| **`name`** | `string` | **Yes** | Unique name of the tool (a-z, 0-9, and underscores only). |
| **`description`** | `string` | **Yes** | Detailed description of what the tool does (used by LLMs for tool selection). |
| **`input`** | `ZodType` | **Yes** | Zod schema defining and validating input arguments. |
| **`auth`** | `{ roles: string[] }` | No | Restricts execution to users with specific roles. |
| **`audit`** | `boolean \| { redactFields: string[] }` | No | Customizes audit logging parameters and redacts PII. |
| **`approval`** | `ApprovalConfig` | No | Suspends execution until human consent is received. |
| **`handler`** | `Function` | **Yes** | The core logic execution function: `(input, context) => Promise<ToolResult>`. |

---

## Example Definition

```typescript
import { defineTool } from '@eagi/sdk';
import { z } from 'zod';

export default defineTool({
  name: 'calculate_tax',
  description: 'Calculates sales tax based on state and amount.',
  input: z.object({
    amount: z.number().positive(),
    state: z.string().length(2)
  }),
  handler: async (input, context) => {
    const { amount, state } = input;
    const taxRate = state === 'CA' ? 0.0825 : 0.05;
    
    return {
      content: [
        {
          type: 'text',
          text: `Sales tax for ${state} on $${amount} is $${(amount * taxRate).toFixed(2)}.`
        }
      ]
    };
  }
});
```

---

## Handler Context (`ToolContext`)

The `handler` function receives the parsed `input` as the first argument, and a `ToolContext` as the second:

*   **`context.identity`**: The user identity propagated by the gateway (`userId`, `role`, OIDC `claims`).
*   **`context.services`**: Initialized services available in the domain workspace.
*   **`context.hooks`**: Access to the hook engine to trigger actions or apply filters.
*   **`context.logger`**: Structured console logger.
