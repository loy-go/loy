---
title: "Layered Clean Architecture"
description: "A beginner-to-advanced guide to Loy's strict 4-layer Clean Architecture model and unidirectional dependency flow."
---

When building Go applications, many developers start with a flat package structure (`models/`, `controllers/`, `views/` or a single `main.go`). While this works for tiny prototypes, it quickly becomes unmaintainable as projects grow:
- Database structs leak into HTTP handlers.
- HTTP request objects leak into database queries.
- Business rules get scattered across handlers, database triggers, and controllers.
- Writing unit tests requires spinning up live database instances or Docker containers.

Loy eliminates this complexity by organizing Go code into **Clean Architecture** (also known as Onion or Hexagonal Architecture). Every feature slice in Loy is partitioned into four decoupled layers with a strict, non-negotiable dependency direction.

---

## The 4-Layer Dependency Direction

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

The core rule of Clean Architecture is **The Dependency Rule**:
> **Dependencies must only point inward toward business logic.** Inner layers know nothing about outer layers.

| Layer | Package Path | Purpose | What It May Import | What It Must NEVER Import |
|---|---|---|---|---|
| **Domain** | `internal/<feature>/domain` | Pure business entities, domain rules, and repository interfaces | Go standard library only | Transport, Infrastructure, external frameworks (`net/http`, `database/sql`, `gorm`) |
| **Application** | `internal/<feature>/service` | Business use cases, workflow orchestration | Domain layer, pure utilities | Transport frameworks (`fiber.Ctx`), concrete database adapters |
| **Infrastructure** | `internal/<feature>/repository`, `internal/platform` | Concrete adapters: SQL queries, caching, third-party APIs | Domain layer (to implement its interfaces) | Transport layer |
| **Transport** | `internal/<feature>/transport` | External protocol entrypoints: HTTP, gRPC, WebSockets | Application services, Domain entities | Database/SQL drivers (`database/sql`, `sqlc`) |

---

## Detailed Breakdown of Each Layer

### 1. Domain Layer (`internal/<feature>/domain`)

The Domain layer is the heart of your application. It defines **what your business is**, completely independent of how it is stored or served.

#### Example: `domain/order.go`
```go
package domain

import (
	"errors"
	"time"
)

var (
	ErrOrderNotFound    = errors.New("order not found")
	ErrNegativePrice    = errors.New("order price cannot be negative")
	ErrInvalidStatus    = errors.New("order status is invalid")
)

// Order is a pure domain entity. Notice: no database tags, no web types.
type Order struct {
	ID        int64
	TenantID  string
	Total     float64
	Status    string
	CreatedAt time.Time
}

// Business logic lives on domain methods, not in HTTP handlers!
func (o *Order) Cancel() error {
	if o.Status == "shipped" {
		return errors.New("cannot cancel an order that has already shipped")
	}
	o.Status = "cancelled"
	return nil
}
```

#### Consumer-Owned Repository Interface: `domain/repository.go`
The Domain layer defines the storage contract it requires, but never implements it:
```go
package domain

import "context"

type OrderRepository interface {
	FindByID(ctx context.Context, id int64) (*Order, error)
	Create(ctx context.Context, order *Order) error
	Update(ctx context.Context, order *Order) error
}
```

---

### 2. Application Layer (`internal/<feature>/service`)

The Application layer defines **what your system does**. It coordinates use cases: fetching entities via repository interfaces, calling domain methods, and saving changes.

#### Example: `service/service.go`
```go
package service

import (
	"context"
	"fmt"
	"myapp/internal/order/domain"
)

type OrderService struct {
	repo domain.OrderRepository // Depends on abstract interface, NOT concrete database!
}

func NewOrderService(repo domain.OrderRepository) (*OrderService, error) {
	if repo == nil {
		return nil, fmt.Errorf("order repository is required")
	}
	return &OrderService{repo: repo}, nil
}

func (s *OrderService) CancelOrder(ctx context.Context, orderID int64) (*domain.Order, error) {
	order, err := s.repo.FindByID(ctx, orderID)
	if err != nil {
		return nil, fmt.Errorf("finding order %d: %w", orderID, err)
	}

	// Execute pure domain rule
	if err := order.Cancel(); err != nil {
		return nil, err
	}

	// Persist state change
	if err := s.repo.Update(ctx, order); err != nil {
		return nil, fmt.Errorf("saving cancelled order: %w", err)
	}

	return order, nil
}
```

---

### 3. Infrastructure Layer (`internal/<feature>/repository`)

The Infrastructure layer implements the abstract interfaces defined by the Domain layer using concrete libraries (`pgx`, `sqlc`, `go-redis`, AWS SDK).

#### Example: `repository/pg_adapter.go`
```go
package repository

import (
	"context"
	"database/sql"
	"fmt"
	"myapp/internal/order/domain"
)

type PostgresOrderRepository struct {
	db *sql.DB
}

func NewPostgresOrderRepository(db *sql.DB) (*PostgresOrderRepository, error) {
	if db == nil {
		return nil, fmt.Errorf("database connection pool is required")
	}
	return &PostgresOrderRepository{db: db}, nil
}

// Implements domain.OrderRepository interface!
func (r *PostgresOrderRepository) FindByID(ctx context.Context, id int64) (*domain.Order, error) {
	var o domain.Order
	err := r.db.QueryRowContext(ctx, "SELECT id, tenant_id, total, status, created_at FROM orders WHERE id = $1", id).
		Scan(&o.ID, &o.TenantID, &o.Total, &o.Status, &o.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, domain.ErrOrderNotFound
	}
	if err != nil {
		return nil, err
	}
	return &o, nil
}
```

---

### 4. Transport Layer (`internal/<feature>/transport`)

The Transport layer is responsible for external protocols. It reads HTTP request headers and JSON bodies, parses parameters, calls the Application service, and returns HTTP status codes.

#### Example: `transport/http/handler.go`
```go
package http

import (
	"net/http"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"myapp/internal/order/domain"
	"myapp/internal/order/service"
)

type OrderHandler struct {
	svc *service.OrderService
}

func NewOrderHandler(svc *service.OrderService) (*OrderHandler, error) {
	return &OrderHandler{svc: svc}, nil
}

func (h *OrderHandler) Cancel(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid order id"})
	}

	order, err := h.svc.CancelOrder(c.UserContext(), id)
	if err != nil {
		if errors.Is(err, domain.ErrOrderNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(order)
}
```

---

## Life of a Request in Loy

When a user triggers an action (e.g. `POST /api/v1/orders/123/cancel`):

1. **Transport Layer**: The HTTP router matches the route and invokes `OrderHandler.Cancel`. It parses `123` into an `int64`.
2. **Application Layer**: Handler passes `123` into `OrderService.CancelOrder`.
3. **Infrastructure Layer**: Service asks `OrderRepository.FindByID(ctx, 123)`. The PostgreSQL adapter executes SQL and hydrates a pure `domain.Order` struct.
4. **Domain Layer**: Service calls `order.Cancel()`. Business logic verifies order status.
5. **Infrastructure Layer**: Service asks `OrderRepository.Update(ctx, order)` to persist the change.
6. **Transport Layer**: Service returns the updated `*domain.Order` to the handler, which serializes JSON `200 OK`.

Notice that:
- You can unit test `OrderService` in milliseconds with a mock repository in pure Go without database connections.
- You can swap PostgreSQL for SQLite or MongoDB without touching a single line of `domain` or `service` code.
- You can swap Fiber for Chi, Gin, or gRPC without touching business logic.

---

## Automated Enforcement: How Loy Protects You

In typical projects, architecture guidelines are forgotten or violated under deadlines. Loy solves this with **automated compile-time enforcement**:

```bash
loy check
```

If any developer or AI assistant accidentally introduces an illegal import:
```text
ERROR [LOY-ARCH-002] internal/order/domain/order.go:14
  domain package internal/order/domain imports infrastructure package internal/order/repository/pg
  Hint: define repository interface in domain or application layer and implement in infrastructure
```

`loy check` verifies all 15 architectural rules (`ARCH-001` through `ARCH-015`) in milliseconds, ensuring your codebase never rots.
