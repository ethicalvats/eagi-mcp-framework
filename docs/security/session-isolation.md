---
title: "Session Isolation"
description: "How EAGI guarantees data privacy and prevents state leakage in multi-client environments."
---

In enterprise settings, multiple users and agents connect to a shared gateway. By default, standard MCP relays are stateless and lack session tracking, which can lead to data leaks if multiple clients send requests to a shared backend process. 

EAGI solves this by implementing **Session Isolation** at the gateway proxy layer.

---

## The Risk: Stdio Broadcast Leakage

In a simple proxy, responses from a domain server's standard output (`stdout`) are broadcast to all connected clients. If Client A and Client B both send requests with standard auto-incrementing JSON-RPC IDs (e.g. `id = 1`), responses will clash, and Client A could receive sensitive data requested by Client B.

---

## The Solution: Transparent Transaction Rewriting

EAGI intercepts every JSON-RPC request and rewrites the `id` field with a globally unique transaction ID before forwarding it to the domain server.

```
Client Request (ID: 1) -> Gateway -> Rewritten Request (ID: tx-SessionA-17810959) -> Domain
```

### Gateway Transaction Flow

1.  **Incoming Request**: Client A sends a request to `/message` with `id: 1` and `session_id: SessionA`.
2.  **ID Remapping**: The Go Gateway generates a unique key `tx-SessionA-[timestamp]` and saves a mapping in memory:
    ```go
    type originalIDMapping struct {
        SessionID  string      // "SessionA"
        OriginalID interface{} // 1
    }
    ```
3.  **Domain Processing**: The gateway forwards the request with the rewritten ID to the domain's input stream.
4.  **Targeted Response**: When the domain process prints the response to `stdout`, the gateway catches it, extracts `tx-SessionA-[timestamp]`, retrieves the origin mapping, restores the ID to `1`, and forwards the response **only** to the SSE socket associated with `SessionA`.

---

## Subscription Notification Routing

A similar isolation pattern applies to resource subscriptions:
*   When Client A calls `resources/subscribe` for `tasks://active`, the gateway intercepts the request and registers the subscription: `tasks://active -> [SessionA]`.
*   When the domain server updates the resource, it emits a `notifications/resources/updated` notification.
*   The gateway catches the notification on `stdout`, looks up the subscriber map, and relays it **only** to subscribed sessions, keeping unsubscribed sessions isolated.
