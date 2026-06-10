---
title: "Local STDIO Servers"
description: "How the EAGI Go Gateway runs and communicates with local TypeScript domain servers."
---

By default, the EAGI Go Gateway operates in **Local Mode**. When started, it scans the `domains/` folder, spawns a separate background process for each domain, and routes messages using standard input/output (stdio) streams.

---

## Process Lifecycle

For every directory found under `domains/`, the gateway's **Process Manager** executes the following cycle:

1.  **Launch**: Spawns a Node.js process running the SDK:
    ```bash
    npx eagi serve-domain --domain [domain-name]
    ```
2.  **Stdio Pipes**: Hooks into the child process's standard streams:
    *   **Stdin (Write)**: Writes client JSON-RPC requests to the process's input.
    *   **Stdout (Read)**: Continuously scans process output for JSON-RPC response payloads.
    *   **Stderr (Log)**: Pipes Node.js console output and errors to the gateway's console logger.
3.  **Monitoring & Restart**: If a local domain process crashes, the process manager intercepts the exit code and automatically restarts the process after a 2-second cooldown.

---

## Tool Discovery

When booting, the gateway dynamically inspects the folder structures (e.g. `domains/math/tools/`). It automatically discovers tool definitions based on file names (e.g. `add.ts` maps to tool name `add`).

The discovered tools are registered in the gateway's routing table. When a client makes a tool call, the gateway knows precisely which local domain process to send it to.
