---
title: "Define Resources"
description: "Reference guide for exposing queryable data streams and resource templates using the EAGI SDK."
---

Resources represent data streams or files exposed to LLM agents. Define them as standard files under `domains/[domain]/resources/` and export as default imports using the `defineResource` helper function.

---

## Resource Definition Schema

A resource configuration accepts the following properties:

| Property | Type | Required | Description |
| :--- | :--- | :--- | :--- |
| **`uri`** | `string` | **Yes** | Unique URI identifier (e.g. `tasks://active` or template `tasks://{id}`). |
| **`name`** | `string` | **Yes** | Short user-facing name for the resource. |
| **`description`** | `string` | **Yes** | Detailed description (tells the LLM when to read this resource). |
| **`mimeType`** | `string` | **Yes** | Mimetype of the contents (e.g. `application/json` or `text/plain`). |
| **`auth`** | `{ roles: string[] }` | No | Restricts read access to specific roles. |
| **`handler`** | `Function` | **Yes** | Logic execution function: `(params, context) => Promise<string>`. |

---

## Static Resources vs Resource Templates

EAGI automatically inspects the `uri` property of the resource:

*   **Static Resource**: If the `uri` is static (e.g. `billing://terms`), the resource is exposed in the standard `ListResources` array.
*   **Resource Template**: If the `uri` contains wildcard variables (e.g. `tasks://{id}`), the resource is exposed as a template. When Client reads `tasks://123`, the gateway re-maps this call, parses `{ id: "123" }`, and supplies it to the handler.

---

## Example Definition (Resource Template)

```typescript
import { defineResource } from '@eagi/sdk';

export default defineResource({
  uri: 'users://{userId}/status',
  name: 'User Status',
  description: 'Detailed live status for a specific user ID.',
  mimeType: 'application/json',
  handler: async (params, context) => {
    const { userId } = params; // Automatically parsed from URI
    const db = context.services.database;
    
    const status = await db.query('SELECT status, last_active FROM users WHERE id = ?', [userId]);
    return JSON.stringify(status);
  }
});
```
