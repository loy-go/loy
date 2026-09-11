# Loy — Product Requirements Document

**Version:** 1.0  
**Status:** Locked  
**Navigation:** [← Index](./00-INDEX.md) | [Technical Design Specification →](./02-TDS.md)

---

## 1. Product Vision

Loy is a Go developer platform that provides a Laravel-like developer experience through conventions, architecture enforcement, typed code generation, ecosystem integration, runtime conventions, and workspace tooling while remaining idiomatic Go.

## 2. Problem

Go has excellent individual libraries and tools but leaves project structure, architecture enforcement, scaffolding, integration conventions, and developer workflow highly fragmented. Teams repeatedly solve the same seams themselves.

## 3. Product Goal

Provide a cohesive opinionated layer over mature Go tooling without replacing the Go ecosystem or hiding the application's runtime architecture.

## 4. Target Users

- Go developers building APIs and services.
- Teams building fullstack Go applications.
- Teams managing monorepos and multiple Go modules.
- Teams that need enforceable architecture and repeatable scaffolding.

## 5. Core Value Proposition

Loy should make the common path easy while keeping the uncommon path possible.

## 6. Product Principles

- Build-time first.
- Generated code is ordinary Go.
- Explicit composition.
- Architecture is enforceable.
- Mature ecosystem over reinvention.
- Progressive complexity.
- Deterministic generation.
- Safe file ownership.

## 7. Core Capabilities

### Project creation

`loy new`, `loy init`, presets, project discovery, manifest handling.

### Generation

Feature, model, service, repository, handler, request, resource, job, event, listener, policy, CRUD and test generation.

### Validation

`loy check`, dependency graph analysis, layer enforcement, package boundaries, workspace rules.

### Development

`loy dev`, `loy build`, `loy test`, `loy migrate`, `loy queue`.

### Diagnostics

`loy doctor`, structured diagnostics, JSON output, deterministic exit behavior.

### Fullstack

Fiber, Templ, HTMX, optional Vite orchestration.

### Integrations

PostgreSQL/sqlc, Valkey, Asynq, OpenTelemetry, gRPC, Docker, Kubernetes, Helm.

## 8. MVP Requirements

The MVP must create a real Go application that can be generated, checked, tested, built, run, and deployed without requiring Loy at runtime.

## 9. Success Criteria

The canonical acceptance path is:

```text
loy new testapp
cd testapp
loy make feature users
loy check
go test ./...
go build ./...
```

## 10. Non-Goals

Loy is not a replacement for:

- Go toolchain
- ORM ecosystem
- queue systems
- frontend package managers
- Kubernetes
- deployment platforms

It is not a reflection-based framework or runtime application container.

## 11. Future Direction

Plugins, broader integrations, additional architecture profiles, advanced managed-file merging, and richer deployment automation remain post-MVP.

---

**Next:** [02-TDS.md — Technical Design Specification](./02-TDS.md)
