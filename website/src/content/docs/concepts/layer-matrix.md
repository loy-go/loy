---
title: "The Architectural Layer Matrix"
description: "Authoritative specification of Clean Architecture layer boundaries, permitted dependency directions, and platform purity governance."
---

Loy standardizes Clean Architecture for Go backends. Package responsibilities and permissible import directions are strictly governed by **Spec 05** and enforced at build time via `loy check` ([ADR-016](/loy/adrs/)).

---

## 1. The 4-Layer Dependency Hierarchy

In Loy, every feature slice is partitioned across four decoupled layers with a strict, non-negotiable dependency flow:

```text
┌─────────────────────────────────────────────────────────────────────────┐
│                             TRANSPORT LAYER                             │
│   HTTP Handlers (Fiber / Chi / Gin / net/http) • gRPC • WebSockets      │
│   Package: internal/<feature>/transport/http, grpc, ws                  │
└────────────────────────────────────┬────────────────────────────────────┘
                                     │ Calls application use cases
                                     ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                            APPLICATION LAYER                            │
│           Use Cases • Service Orchestration • Transaction Flow          │
│           Package: internal/<feature>/service, command, query           │
└──────────────────┬───────────────────────────────────┬──────────────────┘
                   │ Manipulates pure entities         │ Consumes interfaces
                   ▼                                   ▼
┌──────────────────────────────────────┐    ┌─────────────────────────────┐
│             DOMAIN LAYER             │    │    INFRASTRUCTURE LAYER     │
│  Pure Entities • Domain Logic        │    │  PostgreSQL (pgx/sqlc)      │
│  Repository Interfaces • Domain Errs │◀───┤  Valkey Cache • Asynq Queue │
│  Package: internal/<feature>/domain  │    │  Package: repository, plat. │
│  (100% Pure Go Standard Library)     │    │  (Implements Domain Intfs)  │
└──────────────────────────────────────┘    └─────────────────────────────┘
```

The core axiom of Clean Architecture is **The Dependency Rule**:
> **Dependencies must point strictly inward toward pure business logic.** Inner layers know nothing about outer layers.

---

## 2. Permitted Import Matrix

The table below defines the authoritative import matrix enforced by `loy check` (`ARCH-002` through `ARCH-010`):

| Calling Layer | May Import Into | Strictly Prohibited Imports | Governing Rules |
|---|---|---|---|
| **Domain** | Domain, Pure Platform | Infrastructure, Transport, DB drivers (`database/sql`, `pgx`), Web frameworks | `ARCH-002`, `ARCH-003`, `ARCH-008` |
| **Application** | Domain, Application, Pure Platform | Transport (`fiber.Ctx`, `gin.Context`), Concrete Infrastructure adapters | `ARCH-004`, `ARCH-005`, `ARCH-009` |
| **Infrastructure** | Domain (to implement interfaces), Infrastructure, Platform | Transport layer | `ARCH-006` |
| **Transport** | Application (services), Domain (entities), Transport, Platform | Database drivers, SQL queries (`sqlc`, `pgx`) | `ARCH-007` |
| **Platform (`pkg/*`)** | Platform, Standard Library | Domain, Application (no upward imports) | `ARCH-015` |

---

## 3. Detailed Layer Responsibilities & Code Patterns

### 3.1 Domain Layer (`internal/<feature>/domain`)
The domain layer represents core enterprise business rules. It is completely isolated from web protocols, databases, and third-party SDKs.

- **Characteristics**:
  - Contains entities, value objects, domain errors, and repository interfaces.
  - Zero external module imports; pure Go standard library only (`time`, `errors`, `fmt`).
  - Contains rich business methods on entity structs.

#### Compliant Pattern: Pure Entity with Methods
```go title="internal/order/domain/order.go"
package domain

import (
	"errors"
	"time"
)

var (
	ErrOrderNotFound = errors.New("order not found")
	ErrNegativeTotal = errors.New("order total cannot be negative")
)

type Order struct {
	ID        int64
	TenantID  string
	Total     float64
	Status    string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (o *Order) Complete() error {
	if o.Status == "cancelled" {
		return errors.New("cannot complete a cancelled order")
	}
	o.Status = "completed"
	o.UpdatedAt = time.Now().UTC()
	return nil
}
```

---

### 3.2 Application Layer (`internal/<feature>/service`, `command`, `query`)
The application layer coordinates business use cases. It orchestrates domain entities, enforces transactional boundaries, and delegates persistence through interfaces.

- **Characteristics**:
  - Implements application use cases (`service.CreateOrder`, `command.CheckoutCartHandler`).
  - Never references HTTP concepts (`*http.Request`, `fiber.Ctx`, status codes).
  - Depends only on Domain contracts; never directly imports database adapters (`repository.PostgresRepository`).

#### Compliant Pattern: Consumer-Owned Interface Injection
```go title="internal/order/service/service.go"
package service

import (
	"context"
	"fmt"

	"github.com/example/bookstore/internal/order/domain"
)

type Service struct {
	repo domain.OrderRepository
}

func NewService(repo domain.OrderRepository) (*Service, error) {
	if repo == nil {
		return nil, fmt.Errorf("repository dependency is required")
	}
	return &Service{repo: repo}, nil
}

func (s *Service) CompleteOrder(ctx context.Context, id int64) error {
	order, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("finding order: %w", err)
	}

	if err := order.Complete(); err != nil {
		return fmt.Errorf("completing order: %w", err)
	}

	return s.repo.Update(ctx, order)
}
```

---

### 3.3 Infrastructure Layer (`internal/<feature>/repository`, `internal/platform`)
The infrastructure layer implements the interfaces defined by inner layers, communicating with databases, caches, messaging queues, and third-party APIs.

- **Characteristics**:
  - Implements repository interfaces defined in `domain`.
  - Encapsulates SQL queries (`sqlc`), connection pooling (`pgxpool.Pool`), and Redis/Valkey commands.
  - Never leaks database query models into application or transport layers.

#### Compliant Pattern: PostgreSQL Repository Adapter
```go title="internal/order/repository/pg_adapter.go"
package repository

import (
	"context"
	"fmt"

	"github.com/example/bookstore/internal/order/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresRepository(pool *pgxpool.Pool) (*PostgresRepository, error) {
	if pool == nil {
		return nil, fmt.Errorf("database pool is required")
	}
	return &PostgresRepository{pool: pool}, nil
}

func (r *PostgresRepository) FindByID(ctx context.Context, id int64) (*domain.Order, error) {
	query := `SELECT id, tenant_id, total, status, created_at, updated_at FROM orders WHERE id = $1`
	var o domain.Order
	err := r.pool.QueryRow(ctx, query, id).Scan(&o.ID, &o.TenantID, &o.Total, &o.Status, &o.CreatedAt, &o.UpdatedAt)
	if err != nil {
		return nil, domain.ErrOrderNotFound
	}
	return &o, nil
}
```

---

### 3.4 Transport Layer (`internal/<feature>/transport/http`, `grpc`, `ws`)
The transport layer translates external protocol requests (HTTP JSON, gRPC protobuf, WebSocket binary frames) into application use cases, and formats responses back to clients.

- **Characteristics**:
  - Binds HTTP routes (`RegisterRoutes(router fiber.Router)`).
  - Handles JSON decoding, request validation, and HTTP status code mapping.
  - Never executes direct SQL queries (`ARCH-007`).

---

## 4. Platform Layer & Purity Governance (`ARCH-015`)

Loy enforces strict purity rules on packages located in `internal/platform/*` or `pkg/*` ([ADR-023](/loy/adrs/)):

1. **Pure Platform Packages**: Contain strictly in-memory computations (crypto hashing, string utilities, math algorithms). These may be imported by any layer, including Domain.
2. **Impure Platform Packages**: Contain external I/O, network communication, database drivers, or OS handles (`net/http`, `database/sql`, `os`). These must **never** be imported into Domain or Application layers.
3. **No Upward Inversion**: Platform packages must never import feature domain entities or application use cases.

---

## 5. Configurable Architectural DAG (`loy.yaml`)

Teams utilizing Hexagonal, CQRS, or custom topologies can define custom directed graph rules in `loy.yaml`:

```yaml title="loy.yaml"
architecture:
  pattern: custom
  strict: true
  layers:
    transport:
      allows: [application, platform]
      match: ["internal/transport/**", "cmd/**"]
    application:
      allows: [domain, platform]
      match: ["internal/application/**", "internal/*/service/**"]
    domain:
      allows: [platform]
      match: ["internal/domain/**", "internal/*/domain/**"]
    infrastructure:
      allows: [domain, platform]
      match: ["internal/infrastructure/**", "internal/*/repository/**"]
```

When `loy check` runs, it builds an adjacency matrix from this configuration and verifies every package import edge against the graph.
