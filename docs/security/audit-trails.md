---
title: "Audit Trails & Approvals"
description: "Maintain compliance with structured log chaining and approval gates for destructive operations."
---

Enterprise deployments require comprehensive auditing and safety guardrails to ensure AI agents do not perform unauthorized or harmful actions.

---

## 1. Structured Audit Logging

EAGI logs every transaction in a machine-readable JSON format, suitable for routing to SIEM tools or centralized stdout log aggregators:

```json
{
  "timestamp": "2026-06-10T18:23:10Z",
  "domain": "billing",
  "tool": "approve_invoice",
  "user_id": "usr-1234",
  "role": "manager",
  "status": 200,
  "duration_ms": 142,
  "metadata": {
    "invoice_id": "inv-990"
  }
}
```

### PII Redaction
Audit logs automatically respect redact instructions defined in tool definitions:
```typescript
export default defineTool({
  name: 'process_payment',
  audit: { redactFields: ['creditCard', 'cvv'] },
  // ...
```
Fields marked for redaction are automatically stripped or hashed before writing logs.

---

## 2. Interactive Approval Gates

For tools representing high-risk or destructive actions, EAGI supports **Interactive Approval gates**:

```typescript
export default defineTool({
  name: 'delete_database',
  approval: {
    required: true,
    message: (input) => `Are you sure you want to delete table ${input.tableName}?`,
    timeout: '5m'
  },
  // ...
```

When an LLM agent attempts to execute this tool:
1.  **Intercept**: The gateway pauses execution and transitions the transaction state to `input_required` / `awaiting_approval`.
2.  **Notification**: The gateway pushes a notification to the client UI.
3.  **User Consent**: The tool handler remains suspended until the human user clicks "Approve" (or "Reject") in the web UI.
4.  **Resumption**: Upon confirmation, the gateway resumes the tool execution, injecting the approval signature into the context.
