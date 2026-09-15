---
title: "Why Loy? Framework & Architecture Comparison"
description: "A comprehensive comparison between Loy, raw Go, heavy monolithic frameworks, ORMs, and AI coding workflows."
---

When engineering backend applications in Go, engineering teams typically face a difficult dilemma:
1. **The Raw Go Approach**: Start from scratch with standard libraries or light routers (`chi`, `gin`). You achieve pure Go binaries, but spend weeks writing repetitive boilerplate: migrations, config parsers, logging, health checks, metrics, and wiring. As the team grows, architectural discipline breaks down into circular dependencies and layer leaks.
2. **The Heavy Framework Approach**: Adopt monolithic frameworks (Buffalo, Beego, or translate habits from Laravel/Rails/Spring). You get fast initial scaffolding, but become trapped in proprietary runtime base classes, dynamic reflection DI, heavy ORMs, and slow upgrades.

Loy provides a third, modern way: **Build-time platform scaffolding and automated architecture enforcement with zero runtime framework dependencies.**

---

## High-Level Comparison Matrix

| Dimension | Raw Go (Gin / Chi / sqlc) | Buffalo / Beego | GORM / Ent | **Loy Developer Platform** |
|---|---|---|---|---|
| **Runtime Dependency** | Zero | Heavy proprietary runtime | Heavy ORM library | **Zero runtime lock-in** ([ADR-002](/loy/adrs/)) |
| **Architecture Enforcement** | Manual PR reviews | None (Convention only) | None | **Automated AST Linting** (`loy check`) |
| **Dependency Injection** | Manual / `wire` reflection | Dynamic Service Locator | Global / Reflection | **Explicit Go Constructors** in `wiring.go` |
| **Persistence Philosophy** | Handwritten SQL scripts | Active Record ORM | Graph DSL / Reflection | **Type-Safe SQLC + Goose** ([ADR-015](/loy/adrs/)) |
| **Multi-Tenancy** | Bespoke WHERE clauses | Custom middleware | Custom query hooks | **PostgreSQL RLS** ([ADR-018](/loy/adrs/)) |
| **Daemon Topography** | Single binary | Monolith HTTP | DB only | **Dual-Daemon (API + Worker Fleet)** |
| **Live Hot Reload** | Third-party (`air`) | Buffalo dev | Custom watchers | **Native Visual TUI** (`loy dev --tui`) |
| **AI Agent Support** | Hallucinated imports | No agent interfaces | Confusing DSLs | **Native MCP Server + Self-Healing** |
| **Code Ownership** | Developer-owned | Framework-owned | Generator-owned | **Explicit Ownership Models** ([ADR-007](/loy/adrs/)) |

---

## Deep Dive: Loy vs. Other Approaches

### 1. Loy vs. Raw Go (Handwritten Boilerplate)

Many senior Go developers pride themselves on writing "pure Go without frameworks." Loy completely agrees with the philosophy of pure Go—in fact, generated Loy applications **are** pure Go.

However, handwritten raw Go projects suffer from predictable scaling pains:

| Challenge | Raw Go Approach | With Loy |
|---|---|---|
| **New Feature Velocity** | An engineer spends 2 hours manually creating `models/`, `repos/`, `handlers/`, SQL migrations, and wiring them in `main.go`. | Run `loy make crud user name:string email:string` to atomically scaffold the entire vertical slice in under 200ms. |
| **Architectural Drift** | Under deadline pressure, a junior engineer or AI agent imports the database driver into a domain entity. Nobody notices until a circular dependency blocks compilation weeks later. | `loy check` and pre-commit hooks run in milliseconds, instantly blocking illegal layer imports before code is committed. |
| **Onboarding** | Every project in the organization invents its own bespoke folder structure and configuration loader. New hires take weeks to learn where things live. | Standard Clean Architecture structure (`domain/`, `service/`, `repository/`, `transport/`) is immediately familiar to any engineer. |

---

### 2. Loy vs. Heavy Monolithic Frameworks (Buffalo / Beego)

Monolithic frameworks attempt to replicate Ruby on Rails or Laravel in Go by creating heavy runtime base packages:

```go
// Traditional Framework (Heavy Runtime Lock-In):
type UserController struct {
    buffalo.BaseController // Trapped in framework runtime!
}
```

#### Why This Fails in Go:
1. **Violates Go Idioms**: Go thrives on narrow interfaces, explicit composition, and simple structs. Forcing inheritance patterns leads to rigid, unidiomatic code.
2. **Framework Abandonment Risk**: If the framework authors abandon the project, your business logic is trapped in an obsolete runtime that cannot upgrade to new Go releases.
3. **Hidden Magic & Runtime Panics**: Frameworks that use dynamic reflection to discover routes and dependencies fail at runtime rather than compile time.

#### How Loy Solves It:
Loy operates **exclusively at build time**. It writes ordinary Go structs that consume standard community libraries (`pgx`, `fiber`, `asynq`, `otel`). If you uninstall the Loy binary, your project continues to compile, test, and deploy with zero changes.

---

### 3. Loy vs. Complex ORMs (GORM / Ent)

Heavy ORMs promise productivity, but introduce steep technical debt in high-scale production systems:
- **Silent N+1 Queries**: Eager loading and lazy loading hide expensive database roundtrips.
- **Untyped Struct Reflection**: ORMs inspect struct tags and field types at runtime, causing allocation overhead and type bugs.
- **Complex Query DSLs**: Writing complex joins or subqueries in an ORM requires learning a pseudo-SQL dialect that is harder to optimize than real SQL.

#### Loy's Philosophy: Seams Over Components ([ADR-001](/loy/adrs/))
Loy standardizes the seam between **Goose** (schema migrations) and **sqlc** (type-safe SQL compilation):
1. You write plain, standard SQL queries in `queries/*.sql`.
2. `sqlc` compiles them into type-safe, sub-millisecond Go methods with zero reflection.
3. You get the raw speed of native PostgreSQL with 100% compile-time type safety.

---

### 4. Loy with AI Coding Agents vs. Raw Go with AI

As code generation shifts from human typing to autonomous agents (Claude Code, Cursor, Copilot, Kilo, Windsurf), Loy provides the **deterministic substrate** that makes agents reliable:

| AI Agent Scenario | Without Loy | With Loy |
|---|---|---|
| **Layer Isolation** | Agent imports `database/sql` into HTTP handlers or domain entities, creating layer leaks. | `loy check --format agent` emits structured self-healing prompts, guiding the agent to self-correct in one shot. |
| **Scaffolding Features** | Agent generates inconsistent folder structures, invents reflection DI containers, and creates syntax errors. | Agent calls `loy_make_crud` over Model Context Protocol (`loy mcp`), generating an atomic, verified 4-layer vertical slice. |
| **Comment Preservation** | Agent accidentally deletes human-written logic or overwrites manual edits. | Loy guards scaffolding inside `// loy:region:...` comment blocks. Corrupted regions trigger rule `ARCH-014` auto-repair. |

---

## When Should You Use Loy?

### Recommended For:
- **Production Web APIs & SaaS Products**: REST, gRPC, and WebSocket backends that require clean architecture, high testability, and multi-tenant security.
- **Microservices & Fleet Architectures**: Standardizing architecture, logging, metrics, and health checks across dozens of services.
- **Teams Using AI Coding Assistants**: Maximizing agent reliability with native MCP tooling and self-healing validation gates.
- **Enterprise Engineering Organizations**: Eliminating bikeshedding over folder structures, linting, and dependency injection.

### Not Recommended For:
- **Single-File Command-Line Utilities**: Tools that fit comfortably in a single `main.go` file.
- **Low-Level Systems Programming**: OS kernels, embedded firmware, or networking drivers that don't need Clean Architecture layers.
