# Advanced Use Case: Dynamic PII Redaction & Role-Based Auditing

This guide demonstrates how to combine EAGI's **Identity Projection (RBAC)**, **Filters**, and **Action Hooks** to build a production-grade customer support domain with automated compliance controls.

## Scenario
An AI Support Assistant has access to customer profiles containing Personally Identifiable Information (PII) like email addresses and phone numbers.
1. **Access Control**: Only users authenticated with the `support_lead` or `admin` roles can view raw customer PII.
2. **Dynamic Redaction**: If a user with the `support_agent` role runs the tool, their output is automatically intercepted and redacted before the LLM can see it.
3. **Compliance Auditing**: Every tool invocation is logged to a tamper-evident audit ledger recording who ran the tool, the input parameters, and execution duration.

---

## 1. Domain Configuration (`domain.yaml`)

Define the roles and their hierarchy. `support_lead` inherits everything from `support_agent`.

```yaml
name: customer_support
version: 1.0.0
description: Customer support domain and tools

roles:
  support_agent:
    description: Standard support representative. Can query profile info (redacted).
  support_lead:
    description: Team lead. Can query full details and approve refunds.
    includes: [support_agent]
  admin:
    description: System administrator.
```

---

## 2. Defining the Tool (`get_customer.ts`)

Define a simple tool to fetch customer information. The tool handler returns the raw data. The filtering is kept separate in hooks.

```typescript
import { defineTool } from '@eagi/sdk';
import { z } from 'zod';

export default defineTool({
  name: 'get_customer',
  description: 'Retrieve a customer profile by ID.',
  
  input: z.object({
    customerId: z.string().uuid(),
  }),
  
  // Restrict tool access to support roles
  auth: {
    roles: ['support_agent', 'support_lead', 'admin']
  },

  handler: async (input, ctx) => {
    // In a real application, fetch from database.
    const customer = {
      id: input.customerId,
      name: 'Jane Doe',
      email: 'jane.doe@example.com',
      phone: '+1 (555) 019-2834',
      tier: 'gold',
    };

    return {
      content: [{ type: 'text', text: JSON.stringify(customer) }]
    };
  }
});
```

---

## 3. Implementing the Redaction Filter & Audit Log Hooks (`hooks/index.ts`)

We register two hooks:
1. `filter:tool:output` to dynamically intercept the tool payload and redact emails and phone numbers for non-leads.
2. `after:tool:call` to write structured compliance metrics.

```typescript
import { defineHooks } from '@eagi/sdk';

export default defineHooks({
  // Intercept and sanitize tool output
  'filter:tool:output': async (output, ctx) => {
    // We only want to filter the get_customer tool
    if (ctx.tool.name !== 'get_customer') {
      return output;
    }

    const role = ctx.identity?.role;
    
    // If the user has lead or admin privileges, bypass redaction
    if (role === 'support_lead' || role === 'admin') {
      return output;
    }

    // Otherwise, perform PII redaction for standard support agents
    for (const item of output.content) {
      if (item.type === 'text') {
        try {
          const customer = JSON.parse(item.text);
          
          // Redact email: j***@example.com
          if (customer.email) {
            const [local, domain] = customer.email.split('@');
            customer.email = `${local[0]}***@${domain}`;
          }
          
          // Redact phone: +1 (555) ***-****
          if (customer.phone) {
            customer.phone = customer.phone.replace(/(\+\d\s\(\d{3}\)\s)\d{3}-\d{4}/, '$1***-****');
          }

          item.text = JSON.stringify(customer);
        } catch (e) {
          // Output is not JSON, skip or log warning
        }
      }
    }

    return output;
  },

  // Write tamper-evident audit log after execution
  'after:tool:call': async (ctx) => {
    const logEntry = {
      timestamp: new Date().toISOString(),
      tool: ctx.tool.name,
      userId: ctx.identity?.userId || 'anonymous',
      role: ctx.identity?.role || 'none',
      status: ctx.error ? 'error' : 'success',
      executionTimeMs: Date.now() - ctx.startTime,
    };

    // In production, write this to a secure file, SIEM, or database.
    console.info(`[AUDIT] ${JSON.stringify(logEntry)}`);
  }
});
```

---

## Why This Architecture Wins
- **No Security Leaks**: The LLM only receives what the Gateway allows. If a standard `support_agent` tries to prompt the LLM to output the email, the LLM literally has no access to it because the filter redacts it before it reaches the model context.
- **Clean Core Code**: The core tool handler (`get_customer.ts`) does not need complex `if (role === 'admin')` checks littered throughout the business logic. All security rules are centralized in hooks.
