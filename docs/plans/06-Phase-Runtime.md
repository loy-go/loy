# Phase 6: Runtime Application Lifecycle Implementation Plan

**Phase:** 6 of 10  
**Status:** Ready for Implementation  
**Estimated Scope:** Application composition scaffold, lifecycle supervisor, graceful shutdown, health probes  
**Primary Specifications:** [08-Runtime-Application-Lifecycle.md](../08-Runtime-Application-Lifecycle.md), [17-Observability.md](../17-Observability.md), [ADR-002](../adrs/ADR-002-minimal-runtime.md), [ADR-003](../adrs/ADR-003-explicit-wiring.md)

---

## 1. Goal & Objectives
Scaffold the standard application runtime structure for generated projects:
- Pure Go runtime with zero Loy CLI dependencies (`ADR-002`).
- Explicit composition root in `internal/app/{app.go,wiring.go}` (`ADR-003`).
- Ordered initialization sequence: Config -> Logger -> Telemetry -> Infrastructure -> Services -> Transport.
- Resilient lifecycle management: context cancellation, signal handling (SIGINT/SIGTERM), graceful teardown within timeout budget.
- Out-of-the-box `/health/live` and `/health/ready` probe endpoints.

---

## 2. Package Architecture of Generated Applications

```text
<generated_app>/
├── cmd/
│   └── api/
│       └── main.go         # Tiny entrypoint calling app.New() and app.Run()
└── internal/
    ├── app/
    │   ├── app.go          # Application struct, lifecycle states, Run/Stop methods
    │   └── wiring.go       # Explicit composition root + managed regions
    ├── config/
    │   └── config.go       # Typed configuration loader (env variables via caarlos0/env)
    └── platform/
        ├── logger/
        │   └── logger.go   # slog initialization with structured JSON/text handlers
        ├── health/
        │   └── health.go   # Checker interface + live/ready HTTP handlers
        └── shutdown/
            └── hook.go     # Graceful teardown coordinator (LIFO cleanup)
```

---

## 3. Concrete Implementation Steps

### Step 6.1: Composition Root Scaffold (`internal/app/wiring.go`)
1. Provide canonical `wiring.go` template containing structured comment markers:
   ```go
   package app

   func (a *App) wireDependencies() error {
       // loy:region:infrastructure
       // loy:endregion

       // loy:region:repositories
       // loy:endregion

       // loy:region:services
       // loy:endregion

       // loy:region:handlers
       // loy:endregion

       // loy:region:routes
       // loy:endregion

       return nil
   }
   ```

### Step 6.2: Lifecycle Coordinator (`internal/app/app.go`)
1. Implement lifecycle states: `Created -> Initializing -> Ready -> Running -> Stopping -> Stopped`.
2. Implement bounded startup:
   - If any step fails during initialization (e.g. database unreachable), execute cleanup hooks for already initialized resources and exit cleanly.
3. Signal handling:
   - Listen for `os.Interrupt` and `syscall.SIGTERM`.
   - Initiate graceful shutdown with configurable timeout (default: 15s).
   - Close HTTP listener, stop worker queues, flush logs and telemetry, close DB pools.

### Step 6.3: Standard Health Probes (`internal/platform/health`)
1. Implement health system:
   - `/health/live`: Returns 200 OK immediately if application process is running (Kubernetes liveness probe).
   - `/health/ready`: Evaluates registered dependencies (e.g. DB ping, Valkey ping). Returns 200 OK only if ready for traffic; otherwise 503 Service Unavailable (Kubernetes readiness probe).

---

## 4. Test Strategy & Acceptance Criteria

### Lifecycle Unit Tests
- Create a test application instance.
- Call `app.Start()`.
- Trigger simulated SIGTERM via context cancellation.
- Verify shutdown completes within 100ms and DB connections close properly.
- Verify startup failure triggers rollback and leaves no leaking goroutines.

### HTTP Probe Tests
- Query `/health/live` -> 200 `{"status":"up"}`.
- Disconnect simulated database -> Query `/health/ready` -> 503 `{"status":"down","details":{"db":"disconnected"}}`.

---

## 5. Definition of Done
- [ ] Composition root template passes `loy check` without warnings.
- [ ] Graceful shutdown verified under high request volume using test harness.
- [ ] Health probe endpoints conform to standard Kubernetes contract.

---

[← Previous: Phase 5 Plan](./05-Phase-Core-Generators.md) | [Back to Plans Index](./README.md) | [Next: Phase 7 Plan →](./07-Phase-Integrations.md)
