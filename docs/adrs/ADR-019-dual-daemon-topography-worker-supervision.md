# ADR-019: Dual-Daemon Topography & Worker Supervision

## Status
Accepted

## Context
Production SaaS applications run asynchronous background tasks and heavy worker processes in separate daemon binaries from the user-facing HTTP server to protect latency and isolate memory load.

## Decision
1. Provide first-class worker binary scaffolding via `cmd/worker/main.go` and `internal/app/worker_wiring.go`.
2. Manage worker task handler registrations via comment region `// loy:region:tasks`.
3. `loy make job <name>` scaffolds background task payloads, processors, and automatically registers handlers into `worker_wiring.go`.
4. `loy dev` supervises both `cmd/api` and `cmd/worker` simultaneously under unified live reload.

## Consequences
- Clean operational separation between HTTP APIs and background task processors.
- Zero manual wiring needed when creating new background tasks.
- Unified developer experience with synchronized hot-reloads.

---

[Back to ADR Index](./README.md) | [Back to Documentation Index](../00-INDEX.md)
