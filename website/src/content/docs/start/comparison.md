---
title: "Why Loy? Framework Comparison"
description: "How Loy compares to raw Go, web frameworks, ORMs, and monolith toolchains."
---

When building backend systems in Go, teams typically face a dilemma: build everything from scratch with manual boilerplate, or adopt heavy, monolithic frameworks that lock the codebase into proprietary reflection and runtime magic.

Loy provides a third way: **Build-time platform scaffolding and architecture enforcement with zero runtime framework dependencies.**

## Comparison Matrix

| Dimension | Raw Go (Gin / Chi / sqlc) | Buffalo / Beego | GORM / Ent | **Loy Platform** |
|---|---|---|---|---|
| **Runtime Dependency** | Zero | Heavy runtime framework | Heavy ORM library | **Zero** ([ADR-002](/loy/adrs/)) |
| **Architecture Enforcement** | Manual PR reviews | None | None | **Automated AST Linting** (`loy check`) |
| **Dependency Injection** | Manual / `wire` reflection | Service Locator | Reflection | **Explicit Go Constructors** ([ADR-003](/loy/adrs/)) |
| **Persistence Philosophy** | SQL scripts | Active Record ORM | Graph DSL / Reflection | **Goose + sqlc** ([ADR-015](/loy/adrs/)) |
| **Multi-Tenancy** | Handcrafted per query | Bespoke middleware | Custom query hooks | **PostgreSQL RLS** ([ADR-018](/loy/adrs/)) |
| **Daemon Topography** | Single binary | Monolith HTTP | DB only | **Dual-Daemon (API + Worker)** |
| **Hot Reload & Dev** | Third-party (`air`) | Buffalo dev | Custom watchers | **Native Supervisor** (`loy dev`) |
| **Code Ownership** | Developer-owned | Framework-owned | Generator-owned | **Explicit Ownership** ([ADR-007](/loy/adrs/)) |

---

## Core Differentiators

### 1. Zero Runtime Framework Lock-In
Generated Loy applications are 100% ordinary Go binaries. They import mature open-source libraries (`fiber`, `pgx`, `asynq`, `golang-jwt`), never a `github.com/loy-go/loy/runtime` package. If you uninstall Loy tomorrow, your application continues to compile and run normally.

### 2. Architecture Enforced at Build Time
Loy's `loy check` parses package import graphs and Go ASTs in milliseconds, blocking layer direction violations and circular dependencies before code reaches production.

### 3. Explicit Constructor Wiring Over Reflection
Instead of magic runtime reflection or runtime service locators that fail on application boot, Loy splices typed standard Go constructors in `internal/app/wiring.go` and `internal/app/wire_<domain>.go`. Compile errors happen at `go build` time.
