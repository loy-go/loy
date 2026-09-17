# Loy — The Build-Time Go Developer Platform

[![Documentation](https://img.shields.io/badge/Documentation-loy--go.github.io%2Floy-blue?style=flat&logo=astro)](https://loy-go.github.io/loy)
[![CI](https://github.com/loy-go/loy/actions/workflows/ci.yml/badge.svg)](https://github.com/loy-go/loy/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/loy-go/loy)](https://goreportcard.com/report/github.com/loy-go/loy)
[![Go Version](https://img.shields.io/badge/Go-1.24+-00ADD8?logo=go)](https://golang.org)
[![License](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](LICENSE)

**Loy** is an opinionated, build-time developer platform and CLI toolchain for Go. It brings a Laravel-like developer experience—typed scaffolding, vertical slicing, and automated architectural boundary enforcement—to Go, while producing ordinary, idiomatic code with **zero runtime framework lock-in**.

---

## Why Loy?

Go developers frequently struggle with two extremes:
1. **The "Flat Package / Wild West" Extreme**: Tiny prototypes start simple, but as teams grow, database structs leak into HTTP handlers, cyclic dependencies multiply, and refactoring becomes perilous.
2. **The "Heavy Framework" Extreme**: Frameworks introduce heavy reflection DI, proprietary ORMs, runtime service locators, and custom context types that lock your codebase into a third-party dependency.

**Loy bridges this gap** by operating exclusively at **build time**:
- **Seams over components**: Loy standardizes the architectural seams between battle-tested libraries (`pgx`, `fiber`/`chi`/`gin`/`net/http`, `asynq`, `goose`, `sqlc`, `otel`) without inventing a proprietary web framework or ORM.
- **Zero runtime dependency**: Generated applications compile as ordinary Go binaries. They do not require the `loy` binary or any runtime Loy module.
- **Explicit constructor injection**: No reflection DI or magic annotations. Dependencies are injected via standard Go constructors in `internal/app/wiring.go` per [ADR-003](docs/adrs/ADR-003-explicit-wiring.md).
- **Strict, verified layer boundaries**: Enforces `Transport -> Application -> Domain <- Infrastructure` via static AST analysis (`loy check`), preventing architecture degradation in CI.

---

## Key Capabilities & Feature Matrix

| Capability | What Loy Provides | Reference Command |
|---|---|---|
| **Clean Architecture Scaffolding** | Full vertical CRUD slices: domain model, repository interface, Postgres adapter, service usecase, HTTP handler, tests, and auto-wiring. | `loy make crud <name> [fields]` |
| **Configurable Architectural DAG** | Directed graph layer validator enforcing unidirectional dependencies, DDD, Hexagonal, CQRS, or custom layer rules. | `loy check` |
| **Fullstack TypeScript Typegen** | Extracts TypeScript interfaces (`types.ts`) and a typed fetch API client (`client.ts`) directly from Go transport DTOs and routes. | `loy gen client` |
| **CQRS & Segregated Pipelines** | Scaffolds command handlers, direct read-optimized queries, and PostgreSQL row-locked idempotency middleware. | `loy make command`, `loy make query`, `loy make idempotency` |
| **Schema-First Ingestion** | Scaffolds complete slices directly from OpenAPI 3.x specifications or live PostgreSQL schemas / SQL DDL scripts. | `loy make from-spec`, `loy make from-db` |
| **Zero-Downtime Safe DDL** | Production database recipes (`index-concurrent`, `shadow-column`) and tri-route PostgreSQL connection pooling (DataPool, SessionPool, DirectConn). | `loy make migration --recipe=...` |
| **Agent-Native AI Platform** | Native Model Context Protocol (MCP) server, Universal Agent rules generator, and self-healing error remediation loops. | `loy mcp`, `loy make agent-rules`, `loy check --fix` |
| **Process Supervision & Live TUI** | Dual-daemon runner supervising API server, background workers, and Vite asset pipeline with debounced hot reload. | `loy dev` |
| **Embedded Database Migrations** | Embedded Goose migration engine executing versioned SQL scripts without external migration tools. | `loy migrate <up\|down\|status>` |
| **Template Customization Engine** | Local template overrides in `.loy/templates/` with two-tier cascade resolution without forking the binary. | `loy make template list`, `loy make template eject` |

---

## Quickstart (Under 60 Seconds)

### 1. Installation

```bash
# Install the latest release binary
go install github.com/loy-go/loy/cmd/loy@latest

# Verify environment and system prerequisites
loy doctor
```

### 2. Scaffold a New Application

```bash
# Create a new Clean Architecture project with Fiber, Postgres, Valkey, and Asynq
loy new bookstore --preset api
cd bookstore
```

Generated project directory layout:
```text
bookstore/
├── cmd/
│   ├── api/main.go               # HTTP server daemon entrypoint
│   └── worker/main.go            # Asynq background worker daemon entrypoint
├── internal/
│   ├── app/
│   │   ├── app.go                # Application composition container
│   │   ├── wiring.go             # Explicit constructor dependency wiring
│   │   └── worker_wiring.go      # Worker task handler registrations
│   ├── config/config.go          # Typed environment configuration
│   ├── platform/
│   │   ├── database/postgres.go  # Tri-route pgx connection pools
│   │   ├── cache/valkey.go       # Valkey/Redis cache adapter
│   │   ├── queue/asynq.go        # Background task queue client
│   │   ├── logger/logger.go      # Structured slog logging
│   │   └── health/health.go      # Liveness & readiness probes
│   └── transport/http/server.go  # HTTP engine lifecycle coordinator
├── loy.yaml                      # Project manifest & capability matrix
├── go.mod
└── go.sum
```

### 3. Generate a Complete Vertical Slice

Scaffold a domain entity, repository interface, Postgres adapter, service, HTTP handler, DTOs, and automatic composition root wiring:

```bash
loy make crud book title:string:required price:float isbn:string:unique
```

### 4. Enforce Architecture Rules

Verify that code strictly obeys architectural layer boundaries, contains no circular dependencies, and adheres to Go senior hygiene standards:

```bash
loy check
```

Auto-remediate layer violations (e.g. extracting consumer interfaces when application layer imports concrete infrastructure):

```bash
loy check --fix
```

### 5. Start Live Development Supervision

Run the application with filesystem monitoring, debounced hot-reloading, and an interactive terminal dashboard:

```bash
loy dev
```

---

## The 4-Layer Architecture Model

Loy enforces the **Dependency Rule**: *dependencies must point strictly inward toward pure business logic*.

```text
┌─────────────────────────────────────────────────────────────────────────┐
│                             TRANSPORT LAYER                             │
│   HTTP Handlers (Fiber / Chi / Gin / net/http) • gRPC • WebSockets      │
└────────────────────────────────────┬────────────────────────────────────┘
                                     │ Calls application use cases
                                     ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                            APPLICATION LAYER                            │
│           Use Cases • Service Orchestration • Transaction Flow          │
└──────────────────┬───────────────────────────────────┬──────────────────┘
                   │ Manipulates pure entities         │ Consumes interfaces
                   ▼                                   ▼
┌──────────────────────────────────────┐    ┌─────────────────────────────┐
│             DOMAIN LAYER             │    │    INFRASTRUCTURE LAYER     │
│  Pure Entities • Domain Logic        │    │  PostgreSQL (pgx/sqlc)      │
│  Repository Interfaces • Domain Errs │◀───┤  Valkey Cache • Asynq Queue │
│  (100% Pure Go Standard Library)     │    │  (Implements Domain Intfs)  │
└──────────────────────────────────────┘    └─────────────────────────────┘
```

| Layer | Package Path | Permitted Imports | Strictly Prohibited Imports |
|---|---|---|---|
| **Domain** | `internal/<feature>/domain` | Go standard library only | Transport, Infrastructure, DB drivers (`database/sql`, `pgx`) |
| **Application** | `internal/<feature>/service`, `command`, `query` | Domain layer, pure utilities | Transport frameworks (`fiber.Ctx`), concrete DB adapters |
| **Infrastructure** | `internal/<feature>/repository`, `internal/platform` | Domain layer (to implement interfaces) | Transport layer (`fiber`, `chi`, `gin`) |
| **Transport** | `internal/<feature>/transport/http`, `grpc`, `ws` | Application services, Domain entities | Database/SQL drivers (`database/sql`, `sqlc`) |

---

## Architectural Rules Catalog (`loy check`)

Loy analyzes package imports and type relationships across 15 automated rules:

- **`ARCH-001`**: Package Dependency Cycles — Prevents import cycles across all packages (Johnson/Tarjan SCC).
- **`ARCH-002`**: Domain Cannot Import Infrastructure — Domain must never import SQL or persistence packages.
- **`ARCH-003`**: Domain Cannot Import Transport — Domain must never import HTTP, gRPC, or WebSocket packages.
- **`ARCH-004`**: Application Cannot Import Transport — Application services must not bind to HTTP contexts.
- **`ARCH-005`**: Application Cannot Import Concrete Infrastructure — Application must depend on consumer interfaces; concrete adapters are injected in `wiring.go`.
- **`ARCH-006`**: Infrastructure Cannot Import Transport — Persistence code must never import HTTP handlers.
- **`ARCH-007`**: Transport Cannot Execute Business Persistence Logic — Handlers must delegate to application services.
- **`ARCH-008`**: Domain Cannot Import External Frameworks — Domain remains pure Go standard library.
- **`ARCH-009`**: Application Cannot Import Web Frameworks — No `fiber.Ctx` or `gin.Context` in use cases.
- **`ARCH-010`**: Unidirectional Layer Flow — Validates topological DAG matching rules.
- **`ARCH-011`**: No Reflection Service Locators — Mandates explicit Go constructor injection.
- **`ARCH-012`**: No Global Mutable Package State — Forbids package-level mutable variables.
- **`ARCH-013`**: Workspace Module Boundary Isolation — Governs multi-module workspace imports.
- **`ARCH-014`**: Managed File & Comment Region Purity — Verifies generated header tags and splicing regions.
- **`ARCH-015`**: Platform Layer Purity — Platform packages must not import business domain entities.

---

## CLI Command Reference

```text
loy [command]

Project & Architecture:
  new         Create a new Loy project with preset configuration (api, fullstack, minimal, saas)
  init        Initialize Loy configuration (loy.yaml) in an existing Go project
  check       Validate architectural rules, layer boundaries, and circular dependencies (--fix)
  graph       Inspect and visualize package dependency graph in Mermaid or Graphviz format
  doctor      Validate development environment, toolchain prerequisites, and runtime config
  dev         Run multi-process live development server with filesystem watcher & TUI

Code Generation (loy make):
  make crud        Scaffold full vertical slice (migration, domain, repo, svc, handler, wiring)
  make model       Scaffold domain entity struct
  make repository  Scaffold repository interface and database adapter
  make service     Scaffold application usecase service
  make handler     Scaffold HTTP transport handler
  make command     Scaffold CQRS command DTO and transactional handler
  make query       Scaffold CQRS read query and handler
  make idempotency Scaffold row-locked idempotency migration and middleware
  make from-spec   Scaffold full Clean Architecture stacks from OpenAPI / JSON Schema
  make from-db     Reverse-engineer models and repositories from PostgreSQL or SQL DDL
  make migration   Scaffold database migration with zero-downtime recipes (raw, index-concurrent, shadow-column)
  make template    Inspect and eject code generator templates (list, eject)
  make agent-rules Generate tailored AI instructions (AGENTS.md, CLAUDE.md, Cursor rules)
  make deploy      Scaffold production Dockerfile, Kubernetes manifests, Helm chart, CI pipeline

Client SDK Typegen:
  gen client  Generate zero-dependency TypeScript types and API client SDK from Go DTOs

Database & Operations:
  migrate     Manage database migrations via embedded Goose engine (up, down, status, redo, reset)
  seed        Execute database seeders
  routes      Inspect and list all registered HTTP and WebSocket endpoints
  mcp         Start native Model Context Protocol (stdio) JSON-RPC 2.0 server for AI agents
  upgrade     Inspect and apply Loy framework updates
  plugin      Manage community plugins (install, list, run)
```

---

## Platform Comparison

| Feature / Metric | Raw Go (Manual) | Heavy Frameworks (Buffalo / Beego) | Web Routers (Gin / Fiber Alone) | Loy Developer Platform |
|---|---|---|---|---|
| **Runtime Dependency** | Zero | High (framework locked) | Medium (router locked) | **Zero (compiles to ordinary Go)** |
| **Dependency Injection** | Manual / wire | Magic reflection DI / tags | None | **Explicit Go constructors (`wiring.go`)** |
| **Architecture Enforcement** | None | Convention only | None | **Automated static AST validator (`loy check`)** |
| **Code Ownership** | 100% Developer | Framework owned | 100% Developer | **Explicit ownership (`Developer` vs `Generated`)** |
| **AI Agent Tooling** | None | None | None | **Built-in MCP Server & DST AST mutators** |
| **Fullstack Typegen** | Third-party setup | Proprietary DSL | None | **Zero-dependency TypeScript client (`loy gen client`)** |
| **Database Migrations** | External binary | Custom ORM | None | **Embedded Goose engine with safe DDL recipes** |
| **Process Supervision** | Air / reflex | Built-in | Air | **Built-in dual-daemon supervisor & TUI (`loy dev`)** |

---

## Documentation & Learning Hub

Comprehensive guides, specifications, and architecture decision records are available on the official documentation site:

- **[Documentation Hub](https://loy-go.github.io/loy)**
- **[Getting Started Guide](https://loy-go.github.io/loy/start/quickstart/)**
- **[Layered Clean Architecture](https://loy-go.github.io/loy/concepts/architecture/)**
- **[Building Vertical CRUD Slices](https://loy-go.github.io/loy/guides/vertical-slice/)**
- **[CQRS & Commands](https://loy-go.github.io/loy/guides/cqrs-and-commands/)**
- **[Fullstack Typegen (TypeScript SDK)](https://loy-go.github.io/loy/guides/typegen-client/)**
- **[Schema-First Ingestion](https://loy-go.github.io/loy/guides/schema-ingestion/)**
- **[Zero-Downtime Database Migrations](https://loy-go.github.io/loy/guides/database-migrations/)**
- **[Model Context Protocol (MCP) Server](https://loy-go.github.io/loy/ai/mcp-server/)**
- **[Architecture Decisions (ADRs)](docs/adrs/README.md)**

---

## Contributing

We welcome contributions from the Go community! Please review [CONTRIBUTING.md](CONTRIBUTING.md) for local development workflows, testing discipline, and the PR review process.

All contributors are expected to follow our [Code of Conduct](CODE_OF_CONDUCT.md).

---

## Security

To report security vulnerabilities, please refer to our [Security Policy](SECURITY.md). Please do not report security issues via public GitHub issues.

---

## License

Loy is open-source software licensed under the [Apache License, Version 2.0](LICENSE).  
Generated application code belongs 100% to you and is free of framework licensing restrictions.
