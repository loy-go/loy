---
title: "ARCH-002–006: Layer Boundaries"
description: "Rules governing unidirectional flow between Transport, Application, Domain, and Infrastructure."
---

## Overview

The 4-layer architecture guarantees that business logic in Domain and Application remains independent of external delivery mechanisms (HTTP/gRPC) and persistence mechanisms (SQL/NoSQL).

---

## ARCH-002: Domain Cannot Import Infrastructure

- **Severity**: `ERROR`
- **Rule**: Packages in `internal/<feature>/domain` must never import packages in `internal/<feature>/repository` or `internal/platform/*`.
- **Why**: Domain models and business entities must remain clean of database details.

### Violation
```go title="internal/user/domain/user.go"
package domain

import "myapp/internal/user/repository" // VIOLATION: Domain imports Infrastructure!
```

### Remediation
Declare an interface in `domain` that the repository package implements:
```go title="internal/user/domain/repository.go"
package domain

type Repository interface {
    FindByID(ctx context.Context, id int64) (*User, error)
}
```

---

## ARCH-003: Domain Cannot Import Transport

- **Severity**: `ERROR`
- **Rule**: Domain packages must never import HTTP, gRPC, or CLI transport layers.

---

## ARCH-004: Application Cannot Import Transport

- **Severity**: `ERROR`
- **Rule**: Application services (`internal/<feature>/service`) must never import HTTP handlers, router contexts, or web framework packages (`fiber.Ctx`, `gin.Context`, `net/http`).
- **Why**: Business use cases should be executable via any transport (HTTP REST, CLI command, Kafka event listener) without modification.

---

## ARCH-005: Application Should Not Import Concrete Infrastructure

- **Severity**: `WARN`
- **Rule**: Application services should depend on Domain repository interfaces rather than concrete Postgres or Valkey struct implementations.

---

## ARCH-006: Infrastructure Cannot Import Transport

- **Severity**: `ERROR`
- **Rule**: Database repositories and message queues must never import HTTP handlers or web transport contexts.
