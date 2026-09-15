---
title: "ARCH-002–006: Layer Boundary Enforcements"
description: "Detailed breakdown of Clean Architecture unidirectional rules governing Transport, Application, Domain, and Infrastructure."
---

The foundation of Clean Architecture is **The Dependency Rule**:
```text
Transport  ──>  Application  ──>  Domain  <──  Infrastructure
```
Inner layers (Domain, Application) contain pure business logic. Outer layers (Transport, Infrastructure) contain ephemeral technology choices (Fiber, Chi, PostgreSQL, Valkey, AWS).

Rules **`ARCH-002` through `ARCH-006`** enforce this unidirectional flow at the package import level.

---

## ARCH-002: Domain Must Not Import Infrastructure
- **Code**: `LOY-ARCH-002`
- **Severity**: `ERROR`
- **Rule**: Packages in `internal/<feature>/domain` must never import `internal/<feature>/repository` or `internal/platform/*`.
- **Rationale**: Domain entities and rules must remain 100% pure Go standard library, completely decoupled from SQL drivers, database tables, and network clients.

### Prohibited Pattern
```go title="internal/order/domain/order.go"
package domain

// VIOLATION: Domain imports concrete persistence package!
import "myapp/internal/order/repository/pg"

type Order struct {
    ID   int64
    Repo *pg.PostgresOrderRepository
}
```

### Permitted Remediation
Invert the dependency! Define a repository interface in `domain`, and implement it in `repository`:

```go title="internal/order/domain/repository.go (Permitted)"
package domain

import "context"

type OrderRepository interface {
    FindByID(ctx context.Context, id int64) (*Order, error)
    Save(ctx context.Context, order *Order) error
}
```

---

## ARCH-003: Domain Must Not Import Transport
- **Code**: `LOY-ARCH-003`
- **Severity**: `ERROR`
- **Rule**: Domain packages must never import HTTP, gRPC, WebSocket, or CLI transport layers.
- **Rationale**: Domain models represent core business concepts, not HTTP JSON payloads or gRPC protobuf messages.

### Prohibited Pattern
```go title="internal/user/domain/user.go"
package domain

// VIOLATION: Domain imports HTTP transport handler/types!
import "myapp/internal/user/transport/http"

type User struct {
    ID   int64
    Req  http.CreateUserRequest // VIOLATION!
}
```

### Permitted Remediation
Keep domain structs pure. Have your transport handlers map incoming request DTOs to domain entities before calling application services:

```go title="internal/user/transport/http/handler.go (Permitted)"
func (h *UserHandler) Create(c *fiber.Ctx) error {
    var req CreateUserRequest
    if err := c.BodyParser(&req); err != nil {
        return c.Status(400).JSON(fiber.Map{"error": err.Error()})
    }

    // Map DTO to pure domain struct
    user := &domain.User{
        Name:  req.Name,
        Email: req.Email,
    }

    return h.svc.RegisterUser(c.UserContext(), user)
}
```

---

## ARCH-004: Application Must Not Import Transport
- **Code**: `LOY-ARCH-004`
- **Severity**: `ERROR`
- **Rule**: Application services (`internal/<feature>/service`) must never import HTTP routers, web framework contexts (`fiber.Ctx`, `gin.Context`, `chi`), or transport handlers.
- **Rationale**: A business use case (e.g. `RegisterUser` or `TransferFunds`) should be executable from an HTTP handler, a gRPC endpoint, a CLI command, or a background Kafka event listener without changing a single line of service code.

### Prohibited Pattern
```go title="internal/order/service/service.go"
package service

// VIOLATION: Application service tightly coupled to Fiber web framework!
import "github.com/gofiber/fiber/v2"

type OrderService struct{}

func (s *OrderService) CreateOrder(c *fiber.Ctx) error { // VIOLATION!
    // Cannot be invoked by a CLI command or background job!
}
```

### Permitted Remediation
Use standard `context.Context` and plain Go types in service signatures:

```go title="internal/order/service/service.go (Permitted)"
package service

import (
    "context"
    "myapp/internal/order/domain"
)

type OrderService struct{
    repo domain.OrderRepository
}

// Universal signature: callable by HTTP, gRPC, CLI, or background jobs!
func (s *OrderService) CreateOrder(ctx context.Context, order *domain.Order) error {
    return s.repo.Save(ctx, order)
}
```

---

## ARCH-005: Application Should Not Import Concrete Infrastructure
- **Code**: `LOY-ARCH-005`
- **Severity**: `WARN` (or `ERROR` in `--strict` mode)
- **Rule**: Application services should depend on abstract Domain repository interfaces, not concrete SQL struct implementations (`*pg.PostgresOrderRepository`).
- **Rationale**: Depending on interfaces allows you to unit test use cases with fast, in-memory mocks without connecting to a real database.

### Prohibited Pattern
```go title="internal/billing/service/service.go"
package service

// VIOLATION: Service depends directly on concrete Postgres struct!
import "myapp/internal/billing/repository/pg"

type BillingService struct {
    repo *pg.PostgresBillingRepository // Hard to mock in unit tests!
}
```

### Permitted Remediation
Accept the Domain interface in constructor:

```go title="internal/billing/service/service.go (Permitted)"
package service

import "myapp/internal/billing/domain"

type BillingService struct {
    repo domain.BillingRepository // Testable with mocks or fakes!
}

func NewBillingService(repo domain.BillingRepository) (*BillingService, error) {
    return &BillingService{repo: repo}, nil
}
```

Wire the concrete adapter into the service in `internal/app/wiring.go`:
```go title="internal/app/wiring.go"
billingRepo, _ := pg.NewPostgresBillingRepository(a.db)
billingSvc, _ := service.NewBillingService(billingRepo)
```

---

## ARCH-006: Infrastructure Must Not Import Transport
- **Code**: `LOY-ARCH-006`
- **Severity**: `ERROR`
- **Rule**: Persistence repositories, database drivers, and message queues must never import HTTP transport handlers.
- **Rationale**: Persistence adapters belong to infrastructure. They only know how to talk to SQL databases, caches, and storage systems.

### Prohibited Pattern
```go title="internal/user/repository/pg_adapter.go"
package repository

// VIOLATION: Database adapter importing web transport!
import "myapp/internal/user/transport/http"

type PostgresUserRepository struct{}

func (r *PostgresUserRepository) Save(c *http.UserContext) error { // VIOLATION!
}
```

### Permitted Remediation
Repositories accept pure domain entities and standard `context.Context`:

```go title="internal/user/repository/pg_adapter.go (Permitted)"
package repository

import (
    "context"
    "myapp/internal/user/domain"
)

func (r *PostgresUserRepository) Save(ctx context.Context, user *domain.User) error {
    // Execute SQL query...
}
```
