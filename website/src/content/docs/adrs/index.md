---
title: "Architectural Decision Records (ADRs)"
description: "Index and summaries of accepted architectural decisions governing the Loy platform."
---

Loy adheres to standard architectural governance. Major technical choices and invariants are formally recorded in Architectural Decision Records (ADRs).

## Accepted ADR Index

| ADR | Title | Key Architectural Principle |
|---|---|---|
| **[ADR-001](/concepts/architecture/)** | Seams over Components | Standardize wiring between mature Go libraries; never build proprietary ORMs or frameworks. |
| **[ADR-002](/concepts/zero-runtime/)** | Minimal Runtime | Generated applications compile as plain Go binaries without runtime framework lock-in. |
| **[ADR-003](/concepts/explicit-wiring/)** | Explicit Wiring | Reject runtime reflection DI; inject dependencies via explicit Go constructors in `wiring.go`. |
| **[ADR-004](/concepts/zero-runtime/)** | Build-Time First | Heavy work occurs at generation and compilation time, ensuring instantaneous production startup. |
| **[ADR-005](/reference/manifest/)** | Capability-Based Integration | Pluggable, interchangeable drivers for HTTP, persistence, caching, and queues. |
| **[ADR-006](/guides/vertical-slice/)** | Typed UI Generation with Templ | Use Templ exclusively for HTML UI rendering; no proprietary template engines. |
| **[ADR-007](/concepts/plan-based-generation/)** | Plan-Based Generation | In-memory atomic generation with journal rollback before touching disk. |
| **[ADR-008](/reference/manifest/)** | Go Workspace Authority | Native integration with `go.work` for multi-application monorepo development. |
| **[ADR-009](/concepts/architecture/)** | Progressive Complexity | Start with single-module monolith; gracefully evolve into microservices or multi-module monorepos. |
| **[ADR-010](/concepts/zero-runtime/)** | Ordinary Go Output | Every generated artifact is idiomatic, clean Go code formatted through `gofmt`. |
| **[ADR-011](/rules/)** | Static Architecture Validation | Automated linting of layer directions and circular dependencies via `loy check`. |
| **[ADR-012](/concepts/architecture/)** | Contextual Design Patterns | Enforce consumer-owned interfaces, domain purity, and no global mutable state. |
| **[ADR-013](/concepts/plan-based-generation/)** | Codegen Engine Selection | Standard Go `text/template` + `gofmt` for Go source; Templ for HTML views. |
| **[ADR-014](/concepts/explicit-wiring/)** | Managed Comment Splicing | Deterministic code injection using comment regions (`// loy:region:...`). |
| **[ADR-015](/guides/database-migrations/)** | Embedded Goose & SQLC Pipeline | Single binary migration management and type-safe query generation. |
| **[ADR-016](/rules/)** | Two-Phase Architecture Engine | Fast AST inspection followed by deep semantic type analysis. |
| **[ADR-017](/guides/live-reload/)** | Process Supervision for loy dev | Multi-process supervisor with live file watcher and graceful reload. |
