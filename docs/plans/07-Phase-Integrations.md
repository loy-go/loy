# Phase 7: Integrations Implementation Plan

**Phase:** 7 of 10  
**Status:** Ready for Implementation  
**Estimated Scope:** Ecosystem capability adapters, Fiber transport, PostgreSQL/sqlc/goose, Valkey, Asynq, OpenTelemetry  
**Primary Specifications:** [07-Integration-Capability-Specification.md](../07-Integration-Capability-Specification.md), [15-Development-Workflow-Toolchain.md](../15-Development-Workflow-Toolchain.md), [ADR-005](../adrs/ADR-005-capabilities.md), [ADR-015](../adrs/ADR-015-database-migration-and-sqlc-pipeline.md)

---

## 1. Goal & Objectives
Deliver Tier 1 production ecosystem integrations:
- **HTTP Transport**: Fiber adapter with standardized middleware (recovery, request ID, slog logger, CORS).
- **Persistence & Migrations**: PostgreSQL driver (`pgx/v5`), embedded `goose` migration engine (`loy migrate`), and `sqlc` query integration.
- **Cache**: Valkey client (`valkey-go`) with connection pooling and typed cache abstractions.
- **Queue**: Asynq client and background worker server with graceful task draining.
- **Observability**: OpenTelemetry SDK integration for distributed tracing and Prometheus metrics.

---

## 2. Package Architecture & Integration Registry

```text
internal/
└── integration/
    ├── registry.go         # IntegrationRegistry mapping capabilities to adapters
    ├── capability.go       # Capability enums: CapHTTP, CapDatabase, CapCache, CapQueue, CapTelemetry
    └── providers/
        ├── http/
        │   ├── fiber.go    # Fiber integration generator & platform templates
        │   └── middleware/ # Recovery, request tracking, OTel HTTP tracing
        ├── database/
        │   ├── postgres.go # pgxpool setup, sqlc config generator, goose runner
        │   └── goose.go    # Embedded goose migration executor
        ├── cache/
        │   └── valkey.go   # Valkey client wrapper template
        ├── queue/
        │   └── asynq.go    # Asynq client, server, and mux router templates
        └── telemetry/
            └── otel.go     # OpenTelemetry tracer/meter provider setup
```

---

## 3. Concrete Implementation Steps

### Step 7.1: Embedded Migration Engine (`loy migrate`)
1. Integrate `github.com/pressly/goose/v3` in-process.
2. Implement CLI commands:
   - `loy migrate up`: Runs pending migrations in `migrations/`.
   - `loy migrate down`: Rollback latest migration batch.
   - `loy migrate status`: Display applied vs pending migrations.
   - `loy migrate create <name>`: Scaffold new SQL migration file (`-- +goose Up` / `-- +goose Down`).
3. Wire database connection resolution from environment variables (e.g. `DATABASE_URL`).

### Step 7.2: PostgreSQL & `sqlc` Scaffolding
1. Emit `sqlc.yaml` configuring `sqlc-gen-go` engine with `pgx/v5`.
2. Provide `internal/platform/database/postgres.go` template configuring `pgxpool.Pool` with health ping and connection limits.
3. Validate presence of `sqlc` CLI during `loy doctor`.

### Step 7.3: Fiber HTTP Transport Adapter
1. Scaffold Fiber engine in `internal/transport/http/server.go`.
2. Standardize middleware stack:
   - Panic recovery with structured stack logging.
   - Trace context propagation + unique Request ID header (`X-Request-ID`).
   - Latency logging via `slog`.

### Step 7.4: Valkey & Asynq Integrations
1. Scaffold Valkey connection in `internal/platform/cache/valkey.go`.
2. Scaffold Asynq background worker engine:
   - Worker server lifecycle managed under application shutdown hooks.
   - Task handler multiplexer registering generated job handlers.

### Step 7.5: OpenTelemetry Integration
1. Scaffold OTel SDK in `internal/platform/telemetry/provider.go`.
2. Connect OTel trace provider to OTLP gRPC/HTTP exporter.
3. Inject tracing middleware into Fiber router and sqlc DB wrapper.

---

## 4. Test Strategy & Acceptance Criteria

### Integration Test Suite (Dockerized)
- Launch test containers for PostgreSQL and Valkey.
- Run `loy migrate up` -> verify database schema tables created.
- Insert test record through sqlc repository -> verify data persisted.
- Write/read cache entry through Valkey adapter -> verify cache hit.
- Enqueue Asynq job -> verify worker consumes and executes task.
- Query API endpoint -> verify OTel span emitted.

---

## 5. Definition of Done
- [ ] `loy migrate` commands execute reliably against real PostgreSQL.
- [ ] Fiber, Valkey, and Asynq platforms initialize and shutdown cleanly.
- [ ] OTel context propagation passes trace context from HTTP into async worker tasks.

---

[← Previous: Phase 6 Plan](./06-Phase-Runtime.md) | [Back to Plans Index](./README.md) | [Next: Phase 8 Plan →](./08-Phase-Developer-Experience.md)
