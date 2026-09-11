# Loy — Technical Implementation Plan

**Version:** 1.0  
**Status:** Locked  
**Navigation:** [← 02 TDS](./02-TDS.md) | [Index](./00-INDEX.md) | [04 Generator Spec →](./04-Generator-Specification.md)

---

## 1. Implementation Sequence

### Phase 1 — Foundation

Implement filesystem abstractions, process runner, diagnostics, CLI root, version command.

### Phase 2 — Project System

Implement project discovery, manifest loading, `go.mod`/`go.work` detection, workspace model, configuration normalization.

### Phase 3 — Generator Engine

Implement generator contract, naming, typed models, renderer abstraction, Go source renderer (`text/template` + `gofmt`), plans, conflict detection, formatting and validation.

### Phase 4 — Architecture Engine

Implement two-phase analysis: fast import AST pass (`go/parser`) and deep type analysis (`go/packages`), dependency graph, layer classification, cycle detection, dependency rules, workspace boundaries, suppressions.

### Phase 5 — Core Generators

Implement model, service, repository, handler, request, resource, test, feature and CRUD generators (with managed region comment splicing via ADR-014).

### Phase 6 — Runtime

Implement lifecycle, explicit wiring, signal handling, graceful shutdown, health probes and worker lifecycle.

### Phase 7 — Integrations

Implement Fiber (transport isolated), PostgreSQL/sqlc (with embedded goose via ADR-015), Valkey, Asynq, OpenTelemetry and application Templ conventions.

### Phase 8 — Developer Experience

Implement `dev` (native fsnotify process supervisor via ADR-017), graph visualization, target selection, JSON output and shell completion.

### Phase 9 — Fullstack

Implement HTMX conventions and Vite orchestration.

### Phase 10 — Deployment

Implement Docker first, then Kubernetes/Helm integrations.

## 2. Package Design Rules

- Explicit constructors.
- Small interfaces owned by consumers.
- No global mutable state.
- No service locator.
- No reflection DI.
- CLI handlers orchestrate; they do not implement domain or generator logic.

## 3. Engineering Gates

Every phase must preserve:

```text
format
vet
unit tests
architecture checks
generator golden tests
generated project build/test
```

## 4. Definition of Done

A release candidate must generate a real application, pass `loy check`, compile, pass tests, start correctly, handle SIGTERM, and produce deterministic output.

---

**Next:** [04-Generator-Specification.md — Generator Specification](./04-Generator-Specification.md)
