# Phase 8: Developer Experience (DX) Implementation Plan

**Phase:** 8 of 10  
**Status:** Ready for Implementation  
**Estimated Scope:** `loy dev` process supervisor, `loy doctor`, dependency graph visualizer, shell completion  
**Primary Specifications:** [09-CLI-Developer-Experience.md](../09-CLI-Developer-Experience.md), [14-Error-Diagnostics.md](../14-Error-Diagnostics.md), [ADR-017](../adrs/ADR-017-built-in-process-supervision-for-loy-dev.md)

---

## 1. Goal & Objectives
Elevate developer velocity with first-class local tooling:
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
│   └── logger.go           # Interleaved colored terminal log prefixer (e.g. [api], [vite])
├── doctor/
│   ├── doctor.go           # Diagnostic health checker
│   ├── check_go.go         # Go compiler version, GOROOT, GOWORK
│   ├── check_tools.go      # sqlc, docker, git, vite/npm tool detection
│   └── check_db.go         # Database connection connectivity check
├── graph/
│   ├── visualizer.go       # Graph exporter
│   ├── render_text.go      # ASCII tree renderer
│   ├── render_dot.go       # Graphviz DOT format renderer
│   └── render_mermaid.go   # Mermaid markdown diagram renderer
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
   - Debounce file change events (e.g. 200ms) to avoid rebuild storms during batch edits.
2. Implement Process Tree Supervisor:
   - Launch application binary: `go run ./cmd/api`.
   - Launch background worker if configured: `go run ./cmd/worker`.
   - Interleave stdout/stderr with distinct colored prefixes (`[api]`, `[worker]`).
   - On change detection: send `SIGTERM` to process group, wait up to 3s grace period, send `SIGKILL` if unresponsive, re-trigger build.

### Step 8.2: Environment Diagnostic Checker (`internal/doctor`)
1. Check host prerequisites:
   - Go version `>= 1.22`.
   - Presence of `sqlc` CLI executable in `$PATH`.
   - Presence of `docker` and Docker daemon status.
   - Presence of Node/npm/pnpm/bun if Vite is enabled.
2. Check project state:
   - Valid `loy.yaml` and `go.mod`.
   - Database connectivity if configured.
3. Output clean summary report:
   - `[✓] Go 1.23 installed`
   - `[!] sqlc not found in PATH -> Hint: run 'go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest'`

### Step 8.3: Architecture Graph Visualizer (`internal/graph`)
1. Add `loy graph` command with output format selector (`--format=ascii|dot|mermaid`).
2. Generate visual dependency diagrams showing:
   - Package dependency hierarchy.
   - Architectural layer classifications.
   - Detected cycles or boundary warnings highlighted in red.

---

## 4. Test Strategy & Acceptance Criteria

### Supervisor Tests
- Trigger file change in watcher -> verify child process receives SIGTERM and restarts.
- Send SIGINT to supervisor -> verify all child processes terminate without leaving orphaned zombies.

### Doctor Tests
- Mock missing tool -> verify `loy doctor` exits with code 1 and outputs exact installation hint.
- Valid environment -> verify `loy doctor` exits with code 0.

### Graph Tests
- Compare generated mermaid graph against golden fixture.

---

## 5. Definition of Done
- [ ] `loy dev` manages live-reload smoothly across multiple processes.
- [ ] Zero orphaned processes left behind on abrupt terminal exit.
- [ ] `loy doctor` accurately diagnoses missing toolchain binaries.
- [ ] Shell completion scripts generate valid outputs for bash and zsh.

---

[← Previous: Phase 7 Plan](./07-Phase-Integrations.md) | [Back to Plans Index](./README.md) | [Next: Phase 9 Plan →](./09-Phase-Fullstack.md)
