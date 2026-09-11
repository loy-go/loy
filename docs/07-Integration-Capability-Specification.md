# Loy — Integration & Capability Specification

**Version:** 1.0  
**Status:** Locked  
**Navigation:** [← 06 Test Strategy](./06-Test-Strategy-Acceptance.md) | [Index](./00-INDEX.md) | [08 Runtime Lifecycle Spec →](./08-Runtime-Application-Lifecycle.md)

---

## Capability Model

Core capabilities:

- HTTP
- Database
- Cache
- Queue
- RPC
- Template
- Telemetry
- Assets

Generators request capabilities, not concrete implementations.

## Tier 1 MVP

Fiber, PostgreSQL, sqlc, Templ, Asynq, Valkey, OpenTelemetry.

## Tier 2

gRPC, Vite, Docker.

## Tier 3

Kubernetes, Helm, GORM, additional databases, queues and caches.

## Integration Registry

Registry is explicit and process-scoped. No global mutable registration.

## Fiber

Loy standardizes bootstrap, middleware conventions, routing and lifecycle. Fiber remains authoritative for HTTP mechanics.

## Database & Migrations (PostgreSQL/sqlc/goose)

Preferred persistence flow:

```text
Application → Repository Interface → PostgreSQL Adapter → sqlc → PostgreSQL
```

- Migrations: Embedded `goose` library handles schema migration lifecycle under `loy migrate`.
- Codegen: `sqlc generate` runs via external CLI toolchain validated by `loy doctor`.
- GORM is optional and must not be silently mixed with sqlc.

## Valkey

Used for caching and selected transient state.

## Asynq

Used for background jobs and scheduling where configured. Workers are explicitly registered.

## OpenTelemetry

Optional but preferred observability integration. Domain code remains vendor-agnostic.

## Templ

Two distinct roles are preserved: Loy source generation and application SSR.

## Vite

Loy orchestrates it but does not replace npm/pnpm/yarn/bun.

---

**Related ADRs:**
- [ADR-001: Seams Over Components](./adrs/ADR-001-seams-over-components.md)
- [ADR-005: Capabilities-Oriented Integrations](./adrs/ADR-005-capabilities.md)
- [ADR-015: Database Migration and sqlc Pipeline](./adrs/ADR-015-database-migration-and-sqlc-pipeline.md)

**Next:** [08-Runtime-Application-Lifecycle.md — Runtime & Application Lifecycle Specification](./08-Runtime-Application-Lifecycle.md)
