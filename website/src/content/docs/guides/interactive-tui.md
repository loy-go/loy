---
title: "Interactive Development Dashboard"
description: "Full-terminal visual multi-process supervision and hot reload with loy dev --tui."
---

While standard `loy dev` outputs unified, colorized streaming logs, modern cloud applications often require watching multiple independent services simultaneously: the API server, the background worker, and frontend asset watchers (Vite or Tailwind).

Loy provides an interactive, full-terminal dashboard:
```bash
loy dev --tui
```

---

## The TUI Layout

Running `loy dev --tui` launches an ANSI-powered terminal dashboard that refreshes at a responsive 500ms cadence without flickering or consuming excessive CPU:

```text
┌─ Loy Dev Dashboard [myapp] ──────────────────────────────────────────────┐
│ Processes:                                                               │
│   ● [RUNNING] api       PID: 34120  Restarts: 0                         │
│   ● [RUNNING] worker    PID: 34128  Restarts: 0                         │
│   ● [RUNNING] vite      PID: 34135  Restarts: 0                         │
├─ Activity Log ───────────────────────────────────────────────────────────┤
│ [api]    04:15:02 Ready in 8ms: http://localhost:8080                    │
│ [worker] 04:15:02 Asynq worker fleet ready (concurrency: 10)            │
│ [vite]   04:15:03 Local:   http://localhost:5173/                        │
│ [api]    04:15:10 GET /api/v1/orders 200 OK (2.1ms)                      │
│ [worker] 04:15:12 Processed job "send_notification" in 14ms             │
├──────────────────────────────────────────────────────────────────────────┤
│ Status: Supervised 3 process(es) | Hot Reload: Active                    │
│ Commands: [q] Quit  [r] Restart all  [c] Clear logs                      │
└──────────────────────────────────────────────────────────────────────────┘
```

---

## Interactive Controls & Shortcuts

The dashboard listens to keystrokes non-blockingly while keeping child processes running:

| Key | Action | Description |
|---|---|---|
| `q` or `Ctrl+C` | **Quit** | Initiates graceful LIFO phased shutdown across all child processes and exits cleanly. |
| `r` | **Restart All** | Immediately terminates all running child binaries, re-runs `go build`, and restarts all supervised services. |
| `c` | **Clear Logs** | Empties the rolling activity log buffer pane. |

---

## Automatic Task Discovery

The TUI automatically discovers all project tasks from your `loy.yaml` manifest and workspace layout:

1. **API Server**: Discovers `cmd/api/main.go` or the primary Go entrypoint.
2. **Worker Fleet**: Detects `cmd/worker/main.go` if background queues (`asynq`) are enabled.
3. **Frontend Bundlers**: Automatically supervises `npm run dev` or `pnpm dev` when `package.json` with Vite or Tailwind is present.

---

## Graceful Termination Handling

When you press `q`, Loy executes an orderly, phased shutdown:
1. Stops the HTTP server from accepting new connections.
2. Waits for background workers to finish active job payloads (bounded by a graceful deadline).
3. Closes database connection pools and cache clients safely.
4. Restores your terminal cursor and screen state cleanly.
