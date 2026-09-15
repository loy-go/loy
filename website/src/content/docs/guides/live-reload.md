---
title: "Live Hot Reload & Process Supervision"
description: "Multi-process orchestration, filesystem watching, and instantaneous hot reload with loy dev and loy dev --tui."
---

Modern full-stack and microservice development requires running multiple processes concurrently: the Go application server, a background queue worker fleet, and optionally a frontend asset bundler (Vite or Tailwind).

In traditional Go setups, developers must juggle multiple terminal tabs or install external wrappers like `air`, `nodemon`, or `foreman`.

Loy provides **native multi-process supervision and hot reload built directly into the CLI** via `loy dev` ([ADR-017](/loy/adrs/)), requiring zero third-party tools.

---

## Starting the Dev Supervisor

```bash
# Standard streaming mode (colorized stdout/stderr logs)
loy dev

# Interactive visual dashboard mode (full-terminal TUI)
loy dev --tui
```

---

## What `loy dev` Does Automatically

When you run `loy dev`, it automatically executes a four-step pipeline:

```text
┌──────────────────┐     ┌──────────────────────┐     ┌──────────────────────┐
│  1. Task         │ ──> │  2. Filesystem       │ ──> │  3. Hot Rebuild      │
│     Discovery    │     │     Watch & Debounce │     │     & Phased Reload  │
└──────────────────┘     └──────────────────────┘     └──────────────────────┘
                                                                 │
                                                                 ▼
                                                      ┌──────────────────────┐
                                                      │  4. Log Aggregation  │
                                                      │     & Supervision    │
                                                      └──────────────────────┘
```

### 1. Automatic Task Discovery
Loy inspects your project layout and `loy.yaml` manifest to automatically detect and supervise all required services:
- **Go API Server**: Detects `cmd/api/main.go` or your primary application entrypoint.
- **Background Worker Fleet**: Automatically launches `cmd/worker/main.go` when background queues (`asynq`) are configured.
- **Frontend Asset Bundler**: Detects `package.json` with scripts like `npm run dev` or `pnpm dev` for Vite/Tailwind builds.

### 2. Filesystem Watcher with 200ms Debounce
Loy uses native OS file system events (`fsnotify`) to monitor:
- All Go source files (`*.go`)
- HTML / Templ templates (`*.templ`, `*.html`)
- Manifest configuration (`loy.yaml`)

It employs a **200ms debounce window**: if an editor or AI assistant saves multiple files in rapid succession, Loy bundles the changes and triggers exactly one rebuild, avoiding redundant CPU spikes.

#### Automatically Ignored Paths
To maintain high performance, Loy automatically ignores:
- Version control directories: `.git/`, `.svn/`
- Dependency caches: `vendor/`, `node_modules/`
- Build artifacts & temp files: `bin/`, `dist/`, `tmp/`, `*.exe`
- Test fixtures: `testdata/`

### 3. Graceful Phased Reload
When a change is detected:
1. Loy sends a graceful termination signal (`SIGINT` or `SIGTERM`) to the running binary.
2. Waits for active HTTP requests to complete within a bounded grace window.
3. Compiles the new binary with `go build -o <temp-bin>`.
4. If compilation succeeds, launches the new binary instantaneously.
5. If compilation fails, preserves the previous working instance and streams the compiler error clearly in red.

### 4. Log Aggregation
Stdout and stderr streams from all child processes are captured, prefixed with the process identifier (`[api]`, `[worker]`, `[vite]`), color-coded, and rendered to your terminal in real time.

---

## The Interactive TUI Mode (`loy dev --tui`)

For multi-service architectures, the interactive dashboard mode provides complete visual visibility into process health and performance:

```text
┌─ Loy Dev Dashboard [myapp] ──────────────────────────────────────────────┐
│ Processes:                                                               │
│   ● [RUNNING] api       PID: 41208  Restarts: 0                         │
│   ● [RUNNING] worker    PID: 41215  Restarts: 0                         │
│   ● [RUNNING] vite      PID: 41220  Restarts: 0                         │
├─ Activity Log ───────────────────────────────────────────────────────────┤
│ [api]    Ready in 6ms: http://localhost:8080                             │
│ [worker] Asynq worker fleet ready (concurrency: 10)                     │
│ [vite]   Local: http://localhost:5173/                                  │
│ [api]    GET /api/v1/users 200 OK (1.4ms)                                │
│ [worker] Processed task "send_welcome_email" in 8ms                      │
├──────────────────────────────────────────────────────────────────────────┤
│ Status: Supervised 3 process(es) | Hot Reload: Active                    │
│ Commands: [q] Quit  [r] Restart all  [c] Clear logs                      │
└──────────────────────────────────────────────────────────────────────────┘
```

### Keyboard Shortcuts in TUI Mode:
- **`q` or `Ctrl+C`**: Initiates an orderly, phased LIFO shutdown across all child processes and cleanly restores your terminal.
- **`r`**: Immediately terminates all processes, re-runs `go build`, and restarts all supervised services.
- **`c`**: Clears the rolling activity log buffer pane.

---

## Configuration in `loy.yaml`

You can customize development server options directly in your `loy.yaml` manifest:

```yaml title="loy.yaml"
version: 1
project:
  name: myapp

defaults:
  http: fiber
  database: postgres

dev:
  tui: true                  # Default to TUI mode
  debounce: 250ms            # File change debounce window
  rebuild_paths:
    - internal/
    - cmd/
    - views/
  ignored_paths:
    - testdata/
    - storage/
```

---

## Phased LIFO Shutdown Mechanics

When terminating `loy dev` (via `Ctrl+C` or pressing `q` in the TUI), Loy executes an orderly **LIFO (Last-In, First-Out)** shutdown:

1. **Stop Incoming Traffic**: Sends shutdown signal to HTTP transport handlers so no new requests are accepted.
2. **Drain Background Queues**: Allows worker fleet to finish active job executions up to a configurable deadline.
3. **Close Connections**: Closes database connection pools (`sql.DB` / `pgxpool`), cache clients (`valkey`), and telemetry flushes (`otel`).
4. **Clean Exit**: Releases OS file locks and exits with code `0`.
