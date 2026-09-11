# Phase 7: Integrations Implementation Plan

**Phase:** 7 of 10  
**Status:** Completed  
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
        └── database/
            ├── driver.go   # Extensible DriverRegistry with PostgreSQL (pgx/v5) adapter
            ├── goose.go    # Embedded goose migration runner (up, down, status, create, redo, reset)
            └── sqlc.go     # sqlc.yaml generator and external toolchain runner
```

Platform templates in `internal/generator/builtin/templates/`:
- `platform_postgres.go.tmpl`
- `platform_valkey.go.tmpl`
- `platform_asynq.go.tmpl`
- `platform_otel.go.tmpl`
- `transport_fiber.go.tmpl`

---

## 3. Concrete Implementation Steps

### Step 7.1: Embedded Migration Engine (`loy migrate`)
1. Integrate `github.com/pressly/goose/v3` in-process.
2. Implement CLI commands:
   - `loy migrate up`: Runs pending migrations in `migrations/`.
   - `loy migrate down`: Rollback latest migration batch.
   - `loy migrate status`: Display applied vs pending migrations.
   - `loy migrate create <name>`: Scaffold new SQL migration file (`-- +goose Up` / `-- +goose Down`).
   - `loy migrate redo`: Roll back the most recent migration and re-apply it.
   - `loy migrate reset`: Roll back all database migrations.
   - `loy migrate version`: Print the current database migration version.
3. Wire database connection resolution from environment variables (e.g. `DATABASE_URL`) with fallback to `loy.yaml`.

### Step 7.2: PostgreSQL & `sqlc` Scaffolding
1. Emit `sqlc.yaml` configuring `sqlc-gen-go` engine with `pgx/v5`.
2. Provide `internal/platform/database/postgres.go` template configuring `pgxpool.Pool` with health ping and connection limits.
3. Validate presence of `sqlc` CLI via `SqlcRunner.IsInstalled`.

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
1. Scaffold OTel SDK in `internal/platform/telemetry/otel.go`.
2. Connect OTel trace provider to standard trace exporter and global propagators.

---

## 4. Test Strategy & Acceptance Criteria

### Test Suites
- In-memory filesystem tests for `loy migrate create` with timestamp prefixing.
- Contract tests for capability validation and registry operations.
- CLI execution tests verifying JSON and text output streams.
- Live database runner tests skipping cleanly when `DATABASE_URL` unset.

---

## 5. Definition of Done
- [x] `loy migrate` commands execute reliably against real PostgreSQL and in-memory mock filesystems.
- [x] Fiber, Valkey, and Asynq platform templates initialize and shutdown cleanly.
- [x] OTel context propagation passes trace context in platform templates.

---

[← Previous: Phase 6 Plan](./06-Phase-Runtime.md) | [Back to Plans Index](./README.md) | [Next: Phase 8 Plan →](./08-Phase-Developer-Experience.md)
