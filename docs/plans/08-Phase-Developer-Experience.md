# Phase 8: Developer Experience (DX) Implementation Plan

**Phase:** 8 of 10  
**Status:** Completed  
**Estimated Scope:** `loy dev` process supervisor, `loy doctor`, dependency graph visualizer, shell completion  
**Primary Specifications:** [09-CLI-Developer-Experience.md](../09-CLI-Developer-Experience.md), [14-Error-Diagnostics.md](../14-Error-Diagnostics.md), [ADR-017](../adrs/ADR-017-built-in-process-supervision-for-loy-dev.md)

---

## 1. Goal & Objectives
Improve developer workflow speed with local tooling:
- `loy dev`: Built-in multi-process file watcher and supervisor using `fsnotify` (`ADR-017`).
- `loy doctor`: Environment and dependency sanity checker with actionable remediation hints.
- `loy graph`: Dependency and architecture layer visualizer (text, DOT, Mermaid).
- Shell auto-completion (bash, zsh, fish) and structured `--json` CLI output across all commands.

---

## 2. Package Architecture & Internal Seams

```text
internal/
├── dev/
│   ├── supervisor.go       # Multi-process orchestrator (API, Worker, Frontend)
│   ├── watcher.go          # fsnotify file tree watcher with debounce logic
│   ├── process_tree.go     # Process group and tree tracker (clean SIGKILL fallback)
│   └── logger.go           # Interleaved colored terminal log prefixer (e.g. [api], [worker])
├── doctor/
│   └── doctor.go           # Diagnostic health checker (Go, tools, project validation)
├── graph/
│   └── visualizer.go       # Graph exporter (ASCII, DOT, Mermaid)
└── cli/
    ├── dev.go              # loy dev command
    ├── doctor.go           # loy doctor command
    ├── graph.go            # loy graph command
    └── completion.go       # loy completion command
```

---

## 3. Concrete Implementation Steps

### Step 8.1: Native Dev Process Supervisor (`internal/dev`)
1. Implement file system watcher:
   - Watch project root excluding `.git`, `node_modules`, `tmp`, and binary outputs.
   - Debounce file change events (200ms default) to avoid rebuild storms during batch edits.
2. Implement Process Tree Supervisor:
   - Launch application binary: `go run ./cmd/api` (or discovered tasks).
   - Launch background worker if configured: `go run ./cmd/worker`.
   - Interleave stdout/stderr with distinct colored prefixes (`[api]`, `[worker]`).
   - On change detection or SIGINT: send `SIGTERM` to process group (`-pgid`), wait up to 3s grace period, send `SIGKILL` if unresponsive, re-trigger build.

### Step 8.2: Environment Diagnostic Checker (`internal/doctor`)
1. Check host prerequisites:
   - Go version `>= 1.22`.
   - Presence of `sqlc` CLI executable in `$PATH`.
   - Presence of `docker` and Docker daemon status.
   - Presence of `git`.
2. Check project state:
   - Valid `loy.yaml` and `go.mod`.
3. Output clean summary report or structured JSON (`[]diagnostics.Diagnostic`):
   - `[✓] Go 1.26.3 installed`
   - `[!] sqlc not found in PATH -> Hint: run 'go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest'`

### Step 8.3: Architecture Graph Visualizer (`internal/graph`)
1. Add `loy graph` command with output format selector (`--format=ascii|dot|mermaid`).
2. Generate visual dependency diagrams showing:
   - Package dependency hierarchy.
   - Architectural layer classifications.
   - Detected boundary violations highlighted with distinct styling.

---

## 4. Test Strategy & Acceptance Criteria

### Supervisor & Watcher Tests
- File change in watcher triggers debounced event notification.
- DiscoverTasks correctly identifies `cmd/api` and `cmd/worker`.
- Color prefix logger formats line output without ANSI corruption when `--no-color` is set.

### Doctor Tests
- Missing tools emit structured `Diagnostic` warnings with exact installation hints.
- Valid environments and project manifests pass cleanly.

### Graph Tests
- Compare generated ASCII, DOT, and Mermaid diagrams against expected representations.

---

## 5. Definition of Done
- [x] `loy dev` manages live-reload smoothly across multiple processes.
- [x] Zero orphaned processes left behind on abrupt terminal exit via process groups.
- [x] `loy doctor` accurately diagnoses missing toolchain binaries and manifests.
- [x] Shell completion scripts generate valid outputs for bash and zsh.

---

[← Previous: Phase 7 Plan](./07-Phase-Integrations.md) | [Back to Plans Index](./README.md) | [Next: Phase 9 Plan →](./09-Phase-Fullstack.md)
