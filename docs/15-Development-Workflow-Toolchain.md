# Loy — Development Workflow & Toolchain Specification

**Version:** 1.0  
**Status:** Locked  
**Navigation:** [← 14 Error Diagnostics Spec](./14-Error-Diagnostics.md) | [Index](./00-INDEX.md) | [16 Security Spec →](./16-Security.md)

---

## Principle

Loy orchestrates tools; it does not replace them.

## Go

Authoritative for build, test, modules, workspaces and core formatting.

## Templ

Used for Loy source generation and optionally application SSR.

## sqlc

Preferred SQL-to-Go generator for PostgreSQL persistence.

## Vite

Optional asset/frontend development integration. Package manager remains user-selected.

## Docker

Optional developer and deployment integration. Not required for ordinary Go execution.

## `loy dev`

Supervises processes such as API, worker and Vite. Critical process failures must not leave orphaned children.

## Tool Discovery

`loy doctor` may inspect tool availability and versions but must not silently install tools.

## Local Infrastructure

Development orchestration may provide PostgreSQL and Valkey via Docker or another explicitly configured mechanism.

---

**Related ADRs:**
- [ADR-008: Go Workspace Authority](./adrs/ADR-008-go-workspace-authority.md)
- [ADR-015: Database Migration and sqlc Pipeline](./adrs/ADR-015-database-migration-and-sqlc-pipeline.md)
- [ADR-017: Built-in Process Supervision for loy dev](./adrs/ADR-017-built-in-process-supervision-for-loy-dev.md)

**Next:** [16-Security.md — Security Specification](./16-Security.md)
