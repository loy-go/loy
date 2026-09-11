# Loy — Runtime & Application Lifecycle Specification

**Version:** 1.0  
**Status:** Locked  
**Navigation:** [← 07 Integration Capability Spec](./07-Integration-Capability-Specification.md) | [Index](./00-INDEX.md) | [09 CLI DX Spec →](./09-CLI-Developer-Experience.md)

---

## Runtime Principle

A generated application is ordinary Go and does not require Loy to run.

## Canonical Layout

```text
cmd/api/main.go
internal/app/{app.go,wiring.go}
internal/config/config.go
internal/platform/{database,cache,queue,telemetry}
internal/<feature>/{domain,service,repository,handler}
```

## Composition Root

Dependencies are explicitly constructed in application wiring (`internal/app/wiring.go`). Managed regions use structured comments (`// loy:region:...`) for deterministic generator registration (see [ADR-014](./adrs/ADR-014-managed-code-splicing-via-comment-regions.md)).

## Initialization

```text
config → logger → telemetry → DB → cache → queue → repositories → services → handlers → routes → workers
```

Startup failure must clean up already-created resources.

## Lifecycle States

Created → Initializing → Ready → Running → Stopping → Stopped.

## Shutdown

SIGINT/SIGTERM trigger bounded graceful shutdown. Every long-running goroutine has an owner and cancellation path.

## Health

Default HTTP applications expose `/health/live` and `/health/ready` where appropriate.

## Runtime Prohibitions

No service locator, reflection discovery, runtime generator, hidden component discovery, or global container.

## `loy dev`

Acts as an external process supervisor. Application semantics remain independent of it.

---

**Related ADRs:**
- [ADR-002: Minimal Runtime](./adrs/ADR-002-minimal-runtime.md)
- [ADR-003: Explicit Wiring](./adrs/ADR-003-explicit-wiring.md)
- [ADR-017: Built-in Process Supervision for loy dev](./adrs/ADR-017-built-in-process-supervision-for-loy-dev.md)

**Next:** [09-CLI-Developer-Experience.md — CLI & Developer Experience Specification](./09-CLI-Developer-Experience.md)
