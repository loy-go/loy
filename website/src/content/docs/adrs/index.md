---
title: "Architectural Decision Records (ADRs)"
description: "Index and summaries of accepted architectural decisions governing the Loy platform."
---

Loy adheres to standard architectural governance. Major technical choices and invariants are formally recorded in Architectural Decision Records (ADRs).

## Accepted ADR Index

| ADR | Title | Key Architectural Principle |
|---|---|---|
| **[ADR-001](/loy/concepts/architecture/)** | Seams over Components | Standardize wiring between mature Go libraries; never build proprietary ORMs or frameworks. |
| **[ADR-002](/loy/concepts/zero-runtime/)** | Minimal Runtime | Generated applications compile as plain Go binaries without runtime framework lock-in. |
| **[ADR-003](/loy/concepts/explicit-wiring/)** | Explicit Wiring | Reject runtime reflection DI; inject dependencies via explicit Go constructors in `wiring.go`. |
| **[ADR-004](/loy/concepts/zero-runtime/)** | Build-Time First | Heavy work occurs at generation and compilation time, ensuring instantaneous production startup. |
| **[ADR-005](/loy/reference/manifest/)** | Capability-Based Integration | Pluggable, interchangeable drivers for HTTP, persistence, caching, and queues. |
| **[ADR-006](/loy/guides/vertical-slice/)** | Typed UI Generation with Templ | Use Templ exclusively for HTML UI rendering; no proprietary template engines. |
| **[ADR-007](/loy/concepts/plan-based-generation/)** | Plan-Based Generation | In-memory atomic generation with journal rollback before touching disk. |
| **[ADR-008](/loy/reference/manifest/)** | Go Workspace Authority | Native integration with `go.work` for multi-application monorepo development. |
| **[ADR-009](/loy/concepts/architecture/)** | Progressive Complexity | Start with single-module monolith; gracefully evolve into microservices or multi-module monorepos. |
| **[ADR-010](/loy/concepts/zero-runtime/)** | Ordinary Go Output | Every generated artifact is idiomatic, clean Go code formatted through `gofmt`. |
| **[ADR-011](/loy/rules/)** | Static Architecture Validation | Automated linting of layer directions and circular dependencies via `loy check`. |
| **[ADR-012](/loy/concepts/architecture/)** | Contextual Design Patterns | Enforce consumer-owned interfaces, domain purity, and no global mutable state. |
| **[ADR-013](/loy/concepts/plan-based-generation/)** | Codegen Engine Selection | Standard Go `text/template` + `gofmt` for Go source; Templ for HTML views. |
| **[ADR-014](/loy/concepts/explicit-wiring/)** | Managed Comment Splicing | Deterministic code injection using comment regions (`// loy:region:...`). |
| **[ADR-015](/loy/guides/database-migrations/)** | Embedded Goose & SQLC Pipeline | Single binary migration management and type-safe query generation. |
| **[ADR-016](/loy/rules/)** | Two-Phase Architecture Engine | Fast AST inspection followed by deep semantic type analysis. |
| **[ADR-017](/loy/guides/live-reload/)** | Process Supervision for loy dev | Multi-process supervisor with live file watcher and graceful reload. |
| **ADR-018** | PostgreSQL RLS Multi-Tenancy Engine | Native database-level tenant isolation via PostgreSQL Row Level Security and session variables. |
| **ADR-019** | Dual-Daemon Topography & Worker Supervision | First-class background worker binary (`cmd/worker`) and supervised multi-process fleet under `loy dev`. |
| **ADR-020** | Modular Sub-Domain Composition Roots | Modular `wire_<domain>.go` sub-domain wire methods preventing flat monolithic wiring bloat. |
| **ADR-021** | Transport Seam Expansion (gRPC & WebSockets) | High-performance gRPC Proto services and real-time bidirectional WebSocket hubs. |
| **ADR-022** | Baseline Authentication & Session Kit | Standard Go JWT token minting/verification, bcrypt password hashing, and auth/RBAC middlewares. |
| **ADR-023** | Platform Layer & Purity Governance | `pkg/*` mapped to `LayerPlatform` with strict `ARCH-015` IO/network purity checks. |
