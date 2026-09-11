# Loy — Technical Design Specification

**Version:** 1.0  
**Status:** Locked  
**Navigation:** [← 01 PRD](./01-PRD.md) | [Index](./00-INDEX.md) | [03 TIP →](./03-TIP.md)

---

## 1. System Model

Loy consists of a CLI and a set of build-time systems that understand projects, workspaces, generators, architecture, integrations, configuration, diagnostics, and process orchestration.

## 2. Core Architecture

```text
CLI
 ├── Project / Workspace
 ├── Configuration
 ├── Generator Engine
 ├── Architecture Engine
 ├── Integration Registry
 ├── Process / Tooling
 └── Diagnostics
```

## 3. Generated Application Architecture

```text
Transport → Application → Domain
Infrastructure → implements dependencies required by inner layers
```

## 4. Repository Layout

```text
loy/
├── cmd/loy/
├── internal/
│   ├── cli/
│   ├── config/
│   ├── discovery/
│   ├── workspace/
│   ├── project/
│   ├── manifest/
│   ├── generator/
│   ├── architecture/
│   ├── graph/
│   ├── integration/
│   ├── diagnostics/
│   ├── process/
│   ├── filesystem/
│   ├── preset/
│   └── version/
├── generators/
├── integrations/
├── presets/
├── templates/
├── examples/
├── testdata/
└── docs/
```

## 5. Runtime Model

The generated application has an explicit composition root, constructor injection, lifecycle management, signal handling, graceful shutdown, and explicit worker registration.

## 6. Generation Model

```text
generator input
→ normalize
→ validate
→ context
→ generation model
→ render
→ format
→ validate
→ generation plan
→ conflict detection
→ apply
```

## 7. Integration Philosophy

Loy owns seams, not components. Mature external systems remain authoritative for their own behavior.

## 8. Architecture Enforcement

`loy check` analyzes dependency direction, package boundaries, workspace topology, forbidden imports, generated ownership boundaries, and configured architecture rules.

## 9. Technology Baseline

Go, Fiber, PostgreSQL, sqlc, Valkey, Asynq, gRPC, Templ, HTMX, Vite, OpenTelemetry, slog, Docker, Kubernetes, Helm.

## 10. Key ADRs

Detailed design records are located in [adrs/](./adrs/README.md):
- [ADR-001](./adrs/ADR-001-seams-over-components.md) seams over components
- [ADR-002](./adrs/ADR-002-minimal-runtime.md) minimal runtime
- [ADR-003](./adrs/ADR-003-explicit-wiring.md) explicit wiring
- [ADR-004](./adrs/ADR-004-build-time-first.md) build-time-first
- [ADR-005](./adrs/ADR-005-capabilities.md) capabilities
- [ADR-006](./adrs/ADR-006-typed-source-generation-with-templ.md) typed source generation with Templ (superseded by [ADR-013](./adrs/ADR-013-codegen-engine-selection.md))
- [ADR-007](./adrs/ADR-007-plan-based-generation.md) plan-based generation
- [ADR-008](./adrs/ADR-008-go-workspace-authority.md) Go workspace authority
- [ADR-009](./adrs/ADR-009-progressive-complexity.md) progressive complexity
- [ADR-010](./adrs/ADR-010-ordinary-go-output.md) ordinary Go output
- [ADR-011](./adrs/ADR-011-architecture-validation.md) architecture validation
- [ADR-012](./adrs/ADR-012-contextual-patterns.md) contextual patterns
- [ADR-013](./adrs/ADR-013-codegen-engine-selection.md) Go source codegen engine selection
- [ADR-014](./adrs/ADR-014-managed-code-splicing-via-comment-regions.md) managed code splicing
- [ADR-015](./adrs/ADR-015-database-migration-and-sqlc-pipeline.md) database migration and sqlc pipeline
- [ADR-016](./adrs/ADR-016-two-phase-architecture-enforcement-engine.md) two-phase architecture enforcement
- [ADR-017](./adrs/ADR-017-built-in-process-supervision-for-loy-dev.md) process supervision for loy dev

---

**Next:** [03-TIP.md — Technical Implementation Plan](./03-TIP.md)
