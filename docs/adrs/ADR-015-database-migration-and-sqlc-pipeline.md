# ADR-015: Database Migration and sqlc Pipeline

## Status
Accepted

## Context
Loy relies on PostgreSQL and `sqlc` for compile-time safe data access. `sqlc` generates type-safe Go database queries from SQL schema files, but does not run database schema migrations or maintain schema version tables.

## Decision
1. Embed the `pressly/goose` migration library directly into the Loy CLI to power `loy migrate up`, `down`, `status`, and `create`.
2. Keep `sqlc` external: `loy` orchestrates `sqlc generate` via CLI execution, validating its presence and version via `loy doctor`.
3. For full vertical CRUD generation (`loy make crud <entity>`), the pipeline executes:
   - Scaffold timestamped migration SQL file (`migrations/`).
   - Scaffold query SQL file (`queries/`).
   - Run `sqlc generate` (if tool is available and configured).
   - Scaffold repository interface (domain/application) and sqlc adapter (infrastructure).
   - Scaffold application service, transport handler, and route registration.

## Consequences
- Single-binary zero-install experience for database migrations via embedded goose.
- Standard ecosystem sqlc CLI is honored without bundling heavy sqlc compiler dependencies in-process.
- Clean separation between schema evolution (goose) and query generation (sqlc).

---

[Back to ADR Index](./README.md) | [Back to Documentation Index](../00-INDEX.md)
