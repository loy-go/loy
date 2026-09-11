# Loy — Reference Application Specification

**Version:** 1.0  
**Status:** Locked  
**Navigation:** [← 20 Versioning Upgrade Spec](./20-Versioning-Upgrade-Migration.md) | [Index](./00-INDEX.md) | [22 Traceability Roadmap →](./22-Traceability-Roadmap.md)

---

## Purpose

The reference application is simultaneously an architecture reference, generator fixture, regression test and documentation example.

## Domain Areas

- users
- organizations
- authentication
- authorization
- billing
- notifications

## Integrations

Fiber, PostgreSQL, sqlc, Valkey, Asynq, Templ, HTMX, OpenTelemetry and Docker. Vite may be enabled.

## Representative Flow

```text
HTTP Handler
  ↓
Application Service
  ↓
Domain
  ↓
Repository Interface
  ↓
PostgreSQL Adapter
```

Asynchronous path:

```text
Application → Event → Queue → Worker → Notification Adapter
```

## Acceptance

The reference application must pass `loy check`, `go test ./...`, `go build ./...`, relevant integration tests and runtime lifecycle tests.

## Generator Relationship

Where practical, the reference application should be built from actual Loy-generated artifacts so it proves Loy can reproduce its own preferred architecture.

---

**Next:** [22-Traceability-Roadmap.md — Requirements Traceability & Final Implementation Roadmap](./22-Traceability-Roadmap.md)
