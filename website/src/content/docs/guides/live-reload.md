---
title: "Live Hot Reload Supervisor"
description: "Multi-process orchestration, filesystem watching, and instantaneous hot reload with loy dev."
---

Modern full-stack and microservice development requires running multiple processes concurrently: the Go application server, a background queue worker, and optionally a frontend asset bundler (Vite or Tailwind).

Loy provides native process supervision through `loy dev` ([ADR-017](/loy/adrs/)) without requiring external tools like `air` or `foreman`.

## Starting the Dev Supervisor

```bash
loy dev
```

### What `loy dev` Does Automatically

1. **Task Discovery**: Inspects `loy.yaml` and project files to detect:
   - Go application server (`cmd/api/main.go`).
   - Frontend asset bundlers (`package.json` with Vite/Tailwind).
   - Database/cache services.
2. **Filesystem Watcher**: Listens for filesystem modifications across all `.go` and template files using `fsnotify` with a 200ms debounce window.
3. **Graceful Reload**: When source files change, `loy dev` sends a termination signal to the running binary, awaits graceful exit, compiles the new binary with `go build`, and restarts it immediately.
4. **Log Aggregation**: Prefixes and colorizes stdout/stderr streams from all supervised child processes into a single unified terminal view.
