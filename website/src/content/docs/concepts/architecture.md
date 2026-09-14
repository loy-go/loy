---
title: "Layered Architecture"
description: "The 4-layer clean architecture model and strict unidirectional dependency rules."
---

Loy structures applications according to strict **Clean Architecture** principles. Every feature package is partitioned into four decoupled layers.

## The 4-Layer Dependency Direction

```text
Transport  ──>  Application  ──>  Domain  <──  Infrastructure
```

- **Transport (`internal/<feature>/transport`)**: Handles external protocols (HTTP handlers, gRPC endpoints, CLI commands). Translates DTOs into domain objects.
- **Application (`internal/<feature>/service`)**: Implements business use cases, orchestrates transactions, and coordinates domain entities.
- **Domain (`internal/<feature>/domain`)**: The heart of the business model. Contains entities, value objects, and repository interfaces. **Never imports Transport or Infrastructure.**
- **Infrastructure (`internal/<feature>/repository` & `internal/platform`)**: Implements database adapters, external API clients, message brokers, and telemetry. Adapts to interfaces defined in Domain.

---

## Architectural Rules

1. **Domain Isolation**: `domain/` must never import packages from `transport/`, `service/`, or `repository/`.
2. **Interface Ownership**: Consumer owns the interface. The `domain/` package declares `Repository` interface; `repository/` implements it.
3. **No Circular Dependencies**: Cyclic imports between packages are rejected by the Go compiler and flagged in advance by `loy check`.
4. **No Global Mutable State**: Package-level variables with mutable state (`var db *sql.DB`) are prohibited. All state is passed explicitly via constructor parameters.
