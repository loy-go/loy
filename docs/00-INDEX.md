# Loy — Final Documentation Pack

Version: 1.0  
Status: Locked for implementation  

This package contains the final Loy specification baseline consolidated from the locked architectural decisions and specification documents produced during design.

## Specifications

1. [01-PRD.md](./01-PRD.md) — Product Requirements Document
2. [02-TDS.md](./02-TDS.md) — Technical Design Specification
3. [03-TIP.md](./03-TIP.md) — Technical Implementation Plan
4. [04-Generator-Specification.md](./04-Generator-Specification.md) — Generator Specification
5. [05-Architecture-Rule-Specification.md](./05-Architecture-Rule-Specification.md) — Architecture Rule Specification
6. [06-Test-Strategy-Acceptance.md](./06-Test-Strategy-Acceptance.md) — Test Strategy & Acceptance Specification
7. [07-Integration-Capability-Specification.md](./07-Integration-Capability-Specification.md) — Integration & Capability Specification
8. [08-Runtime-Application-Lifecycle.md](./08-Runtime-Application-Lifecycle.md) — Runtime & Application Lifecycle Specification
9. [09-CLI-Developer-Experience.md](./09-CLI-Developer-Experience.md) — CLI & Developer Experience Specification
10. [10-Configuration-Manifest.md](./10-Configuration-Manifest.md) — Configuration & Manifest Specification
11. [11-Project-Workspace.md](./11-Project-Workspace.md) — Project & Workspace System Specification
12. [12-Code-Generation-Templates.md](./12-Code-Generation-Templates.md) — Code Generation & Template Specification
13. [13-Architecture-Design-Patterns.md](./13-Architecture-Design-Patterns.md) — Application Architecture & Design Pattern Specification
14. [14-Error-Diagnostics.md](./14-Error-Diagnostics.md) — Error & Diagnostics Specification
15. [15-Development-Workflow-Toolchain.md](./15-Development-Workflow-Toolchain.md) — Development Workflow & Toolchain Specification
16. [16-Security.md](./16-Security.md) — Security Specification
17. [17-Observability.md](./17-Observability.md) — Observability Specification
18. [18-Deployment-Infrastructure.md](./18-Deployment-Infrastructure.md) — Deployment & Infrastructure Specification
19. [19-Plugin-Extension.md](./19-Plugin-Extension.md) — Plugin & Extension Specification
20. [20-Versioning-Upgrade-Migration.md](./20-Versioning-Upgrade-Migration.md) — Versioning, Upgrade & Migration Specification
21. [21-Reference-Application.md](./21-Reference-Application.md) — Reference Application Specification
22. [22-Traceability-Roadmap.md](./22-Traceability-Roadmap.md) — Requirements Traceability & Final Implementation Roadmap

## Detailed Phase Implementation Plans

The engineering rollout is decomposed into 10 phase execution plans in [plans/](./plans/README.md):

- [Phase 1: Foundation](./plans/01-Phase-Foundation.md)
- [Phase 2: Project & Workspace System](./plans/02-Phase-Project-System.md)
- [Phase 3: Generator Engine](./plans/03-Phase-Generator-Engine.md)
- [Phase 4: Architecture Engine](./plans/04-Phase-Architecture-Engine.md)
- [Phase 5: Core Generators](./plans/05-Phase-Core-Generators.md)
- [Phase 6: Runtime Application Lifecycle](./plans/06-Phase-Runtime.md)
- [Phase 7: Integrations](./plans/07-Phase-Integrations.md)
- [Phase 8: Developer Experience](./plans/08-Phase-Developer-Experience.md)
- [Phase 9: Fullstack](./plans/09-Phase-Fullstack.md)
- [Phase 10: Deployment & Infrastructure](./plans/10-Phase-Deployment.md)

## Architectural Decision Records (ADRs)

Detailed rationale is recorded in the [ADR Registry](./adrs/README.md):

- [ADR-001](./adrs/ADR-001-seams-over-components.md) — Seams over components
- [ADR-002](./adrs/ADR-002-minimal-runtime.md) — Minimal runtime
- [ADR-003](./adrs/ADR-003-explicit-wiring.md) — Explicit wiring
- [ADR-004](./adrs/ADR-004-build-time-first.md) — Build-time first
- [ADR-005](./adrs/ADR-005-capabilities.md) — Capabilities-oriented integrations
- [ADR-006](./adrs/ADR-006-typed-source-generation-with-templ.md) — Typed source generation with Templ (superseded in part by ADR-013)
- [ADR-007](./adrs/ADR-007-plan-based-generation.md) — Plan-based generation
- [ADR-008](./adrs/ADR-008-go-workspace-authority.md) — Go workspace authority
- [ADR-009](./adrs/ADR-009-progressive-complexity.md) — Progressive complexity
- [ADR-010](./adrs/ADR-010-ordinary-go-output.md) — Ordinary Go output
- [ADR-011](./adrs/ADR-011-architecture-validation.md) — Architecture validation engine
- [ADR-012](./adrs/ADR-012-contextual-patterns.md) — Contextual patterns over dogmatic boilerplate
- [ADR-013](./adrs/ADR-013-codegen-engine-selection.md) — text/template for Go codegen; Templ for HTML SSR
- [ADR-014](./adrs/ADR-014-managed-code-splicing-via-comment-regions.md) — Managed region splicing in wiring.go
- [ADR-015](./adrs/ADR-015-database-migration-and-sqlc-pipeline.md) — Embedded goose runner + external sqlc generation
- [ADR-016](./adrs/ADR-016-two-phase-architecture-enforcement-engine.md) — Two-phase architecture enforcement engine
- [ADR-017](./adrs/ADR-017-built-in-process-supervision-for-loy-dev.md) — Built-in process supervision for loy dev

## Locked Invariants

- Loy owns seams, not components.
- Architecture is enforced, not ceremony.
- Generated applications remain ordinary Go applications.
- Explicit dependency injection; no service locator or reflection DI.
- Domain does not depend on infrastructure.
- Transport does not own business logic.
- Generators produce plans before filesystem mutation.
- Generated and developer-owned code have explicit ownership.
- Generation is deterministic.
- Configuration is typed and validated.
- Go modules/workspaces remain authoritative for Go behavior.
- Integrations are capability-oriented.
- Runtime does not require Loy.
- Secrets remain external to `loy.yaml`.
- Progressive complexity is preferred over unnecessary pattern ceremony.

## Implementation Status

Specification phase: **LOCKED FOR IMPLEMENTATION**.

Future architectural changes should be introduced through ADRs, followed by specification updates and acceptance tests.
