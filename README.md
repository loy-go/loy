# Loy

[![CI](https://github.com/loy-go/loy/actions/workflows/ci.yml/badge.svg)](https://github.com/loy-go/loy/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/loy-go/loy)](https://goreportcard.com/report/github.com/loy-go/loy)
[![License](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/Go-1.24+-00ADD8?logo=go)](https://golang.org)

**Loy** is an opinionated, build-time developer platform and CLI toolchain for Go. It provides a Laravel-like developer experience—typed scaffolding, vertical slicing, and automated architectural boundary enforcement—while producing ordinary, idiomatic Go with **zero runtime framework dependencies**.

---

## Core Invariants

- **Seams over components**: Loy configures and stitches battle-tested Go libraries (`pgx`, `fiber`, `asynq`, `goose`, `sqlc`, `otel`). It does not build a proprietary web framework or ORM ([ADR-001](docs/adrs/ADR-001-seams-over-components.md)).
- **Zero runtime dependency**: Generated applications are plain Go binaries. They compile without the `loy` binary or any proprietary Loy runtime libraries ([ADR-002](docs/adrs/ADR-002-minimal-runtime.md)).
- **Explicit wiring**: No reflection DI, magic annotations, or global service locators. Dependencies are injected via explicit Go constructors in `internal/app/wiring.go` ([ADR-003](docs/adrs/ADR-003-explicit-wiring.md)).
- **Strict layer direction**: `Transport -> Application -> Domain <- Infrastructure`. The domain layer never imports infrastructure or transport packages ([Doc 05](docs/05-Architecture-Rule-Specification.md)).
- **Plan-based atomic generation**: Code generators produce an in-memory plan before mutating disk. Splicing into existing files occurs through managed comment regions (`// loy:region:...`) ([ADR-007](docs/adrs/ADR-007-plan-based-generation.md), [ADR-014](docs/adrs/ADR-014-managed-code-splicing-via-comment-regions.md)).

---

## Quickstart (Under 60 Seconds)

### 1. Installation

```bash
# Install latest release binary
go install github.com/loy-go/loy/cmd/loy@latest

# Verify environment and prerequisites
loy doctor
```

### 2. Create a New Application

```bash
loy new myapp --preset api
cd myapp
```

This scaffolds a clean architecture project structure:
```text
myapp/
├── cmd/api/main.go
├── internal/
│   ├── app/wiring.go
│   ├── config/config.go
│   └── platform/
├── loy.yaml
├── go.mod
└── go.sum
```

### 3. Generate a Vertical Feature Stack

Scaffold a domain entity, repository interface, database adapter, service, HTTP handler, and wire them up automatically:

```bash
loy make crud user name:string email:string:unique
```

### 4. Enforce Architecture Boundaries

Verify that your project adheres to clean layer boundaries, has no circular dependencies, and follows strict import rules:

```bash
loy check
```

### 5. Live Development Server

Start the application with live hot reload and process supervision:

```bash
loy dev
```

---

## CLI Command Overview

| Command | Description |
|---|---|
| `loy new <name>` | Scaffold a new Loy project with preset configuration (`api`, `fullstack`, `minimal`) |
| `loy init` | Initialize Loy configuration (`loy.yaml`) in an existing Go project |
| `loy make crud <name> [fields]` | Scaffold complete vertical slice (migration, domain, repo, svc, handler, wiring) |
| `loy make <artifact> <name>` | Scaffold atomic component (`model`, `service`, `repository`, `handler`, `job`, `event`, etc.) |
| `loy check` | Validate architectural rules, layer boundaries, and circular dependencies |
| `loy graph` | Visualize package dependency hierarchy in Mermaid or Graphviz DOT format |
| `loy doctor` | Validate development environment, Go toolchain, and project prerequisites |
| `loy migrate <up\|down\|status>` | Manage database migrations via embedded goose engine |
| `loy dev` | Run multi-process live development server with filesystem watcher and hot reload |
| `loy version` | Print version, git commit, build date, and platform info |

---

## Architectural Rules (`loy check`)

Loy includes 14 built-in architectural verification rules:

- **ARCH-001**: Detects circular package dependencies.
- **ARCH-002 - ARCH-006**: Enforces strict layer boundaries (`Domain` cannot import `Transport` or `Infrastructure`).
- **ARCH-007 - ARCH-010**: Enforces import policies (no direct DB drivers in domain, no disallowed third-party imports).
- **ARCH-011 - ARCH-014**: Validates package conventions, naming rules, and forbidden global mutable state.

---

## Documentation

Full architectural specifications and decision records are maintained in [`docs/`](docs/00-INDEX.md):

- [Technical Design Specification (TDS)](docs/02-TDS.md)
- [Technical Implementation Plan (TIP)](docs/03-TIP.md)
- [Architecture Rule Specification](docs/05-Architecture-Rule-Specification.md)
- [Architectural Decision Records (ADRs)](docs/adrs/README.md)

---

## Contributing

We welcome community contributions! Please read [CONTRIBUTING.md](CONTRIBUTING.md) for local development instructions, test guidelines, and our PR validation process.

All contributors are expected to adhere to the [Code of Conduct](CODE_OF_CONDUCT.md).

---

## Security

To report security vulnerabilities, please review our [Security Policy](SECURITY.md). Please do not disclose vulnerabilities via public GitHub issues.

---

## License

Loy is licensed under the [Apache License, Version 2.0](LICENSE).
Generated code belongs 100% to you without any framework license restrictions.
