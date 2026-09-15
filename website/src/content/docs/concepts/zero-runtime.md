---
title: "The Zero Runtime Invariant"
description: "Why Loy never introduces runtime framework lock-in, and how your applications remain pure, ordinary Go binaries."
---

Most web frameworks (Ruby on Rails, Laravel, Django, Spring Boot, Buffalo, Beego) act as heavy runtime dependencies. When you build an application with them:
- Every controller must inherit from a framework base class.
- Every model must extend an Active Record ORM base class.
- Dependency injection relies on runtime reflection, magic annotations, or global service locators.
- Your production code cannot run or compile without the framework package installed.

Loy takes the opposite approach: **Build-Time First, Zero Runtime Lock-In** ([ADR-002](/loy/adrs/)).

```text
Traditional Framework Model:
┌─────────────────────────────────────────────────────────┐
│               Your Application Code                     │
├─────────────────────────────────────────────────────────┤
│            Proprietary Framework Runtime                │  ◀── Heavy dependency
│ (Base Controllers, ORM Engines, Reflection Registries)  │      locks you in forever
└─────────────────────────────────────────────────────────┘

The Loy Invariant:
┌─────────────────────────────────────────────────────────┐
│               Your Application Code                     │
├─────────────────────────────────────────────────────────┤
│          Mature Open-Source Community Libraries         │  ◀── Zero Loy runtime!
│     (pgx, fiber/chi, asynq, goose, sqlc, otel)          │      Pure, ordinary Go.
└─────────────────────────────────────────────────────────┘
```

---

## What Does "Zero Runtime" Mean in Practice?

### 1. Pure Go Output
When you run `loy new` or `loy make crud`:
- Every generated file is standard, idiomatic Go (`.go`).
- Generated packages import only standard library packages and battle-tested, un-wrapped open-source libraries (`github.com/jackc/pgx/v5`, `github.com/gofiber/fiber/v2`, `github.com/hibiken/asynq`).
- **Your application never imports `github.com/loy-go/loy`**.

Check your generated `go.mod`:
```go title="go.mod"
module myapp

go 1.24.0

require (
	github.com/gofiber/fiber/v2 v2.52.5
	github.com/jackc/pgx/v5 v5.7.1
	github.com/hibiken/asynq v0.24.1
	github.com/golang-jwt/jwt/v5 v5.2.1
)
// Notice: NO loy runtime dependencies!
```

### 2. No Proprietary Base Classes
In Loy:
- A domain entity is a plain Go struct:
  ```go
  type User struct {
      ID    int64
      Email string
  }
  ```
- An HTTP handler is a plain Go struct:
  ```go
  type Handler struct {
      svc *service.UserService
  }
  ```
- A service is a plain Go struct:
  ```go
  type UserService struct {
      repo domain.UserRepository
  }
  ```

There are no `loy.Model`, `loy.Controller`, or `loy.BaseService` abstractions.

---

## The "Desert Island" Test

To prove that Loy introduces zero framework lock-in, run this thought experiment:

> **What happens if your team uninstalls or deletes the `loy` CLI binary tomorrow?**

1. **Compilation**: Your project continues to compile instantaneously with standard `go build ./...`.
2. **Testing**: Your test suites continue to run with standard `go test -v ./...`.
3. **Deployment**: Your Dockerfiles and Kubernetes manifests build and run unchanged.
4. **CI/CD**: Your deployment pipelines do not need the `loy` binary installed.
5. **Onboarding**: A new Go engineer who has never heard of Loy can clone the repository, open it in VS Code, and start coding immediately without learning a proprietary framework syntax.

Loy acts as your **build-time velocity multiplier**, not your runtime jailer.

---

## The Hidden Costs of Runtime Framework Lock-In

Why did we make Zero Runtime a non-negotiable architectural invariant?

| Risk | Traditional Runtime Frameworks | Loy Build-Time Architecture |
|---|---|---|
| **Framework Abandonment** | If the framework maintainer stops updating, your entire business code is stranded on an obsolete runtime. | You own 100% of the code. You depend only on Go and mature, active ecosystem libraries. |
| **Upgrade Gridlock** | Moving to new Go language versions often breaks framework reflection internals and monkey-patched runtimes. | Standard Go code upgrades smoothly with standard `go get -u` and `go fix`. |
| **Performance Overhead** | Heavy runtime reflection, dynamic dispatch, and complex middleware chains inflate memory usage and latency. | Zero reflection allocations on startup. Blazing-fast compiled machine code. |
| **Hiring Friction** | Engineers hesitate to learn proprietary framework dialects that don't translate to other companies. | Standard Clean Architecture Go code that every senior engineer recognizes instantly. |
| **Debugging Complexity** | Stack traces are 50 layers deep in framework internals, making root-cause analysis frustrating. | Stack traces are shallow and transparent: Handler -> Service -> Repository -> Database. |

---

## Build-Time Scaffolding vs. Runtime Magic

Instead of resolving dependencies dynamically at runtime with reflection, Loy generates **explicit Go constructors** in `internal/app/wiring.go`:

```go title="internal/app/wiring.go"
func (a *App) wireDependencies() error {
    userRepo, err := userRepo.NewPostgresRepository(a.db)
    if err != nil {
        return err
    }
    userSvc, err := userService.NewService(userRepo)
    if err != nil {
        return err
    }
    userHandler, err := userHttp.NewHandler(userSvc)
    if err != nil {
        return err
    }
    userHandler.RegisterRoutes(a.router)
    return nil
}
```

If a dependency type changes, **the standard Go compiler flags the error at compile time** with an exact line number. There are no runtime panics when your server boots up in staging or production.
