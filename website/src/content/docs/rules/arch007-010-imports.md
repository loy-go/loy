---
title: "ARCH-007–010: Import Restrictions & Framework Purity"
description: "Rules restricting raw database persistence in handlers, framework dependencies in domain/application, and strict layer direction enforcement."
---

Rules **`ARCH-007` through `ARCH-010`** govern framework isolation and layer responsibility boundaries. They prevent framework code (HTTP routers, ORMs, SQL drivers) from leaking into core business logic and stop transport handlers from executing raw database queries directly.

---

## ARCH-007: Transport Must Not Contain Raw Persistence
- **Code**: `LOY-ARCH-007`
- **Severity**: `ERROR`
- **Rule**: Transport files in `internal/.../transport` must never import persistence packages (`database/sql`, `gorm.io/gorm`, or packages containing `sqlc`).
- **Rationale**: Transport handlers exist solely to parse incoming HTTP/gRPC requests, validate basic format, call Application services, and serialize responses. Executing raw SQL queries directly in an HTTP handler bypasses domain validation, authorization policies, transaction coordination, and telemetry.

### Prohibited Pattern
```go title="internal/order/transport/http/handler.go"
package http

import (
    "database/sql" // VIOLATION: Transport handler importing database driver!
    "github.com/gofiber/fiber/v2"
)

type OrderHandler struct {
    db *sql.DB // Bypassing Application use cases!
}

func (h *OrderHandler) GetOrders(c *fiber.Ctx) error {
    // Raw database query in an HTTP handler!
    rows, err := h.db.Query("SELECT id, total FROM orders")
    // ...
}
```

### Permitted Remediation
Move database queries into an Infrastructure repository and coordinate them through an Application service use case:

```go title="internal/order/transport/http/handler.go (Permitted)"
package http

import (
    "github.com/gofiber/fiber/v2"
    "myapp/internal/order/service"
)

type OrderHandler struct {
    svc *service.OrderService // Clean: delegates business operations to use cases!
}

func (h *OrderHandler) GetOrders(c *fiber.Ctx) error {
    orders, err := h.svc.ListOrders(c.UserContext())
    if err != nil {
        return c.Status(500).JSON(fiber.Map{"error": err.Error()})
    }
    return c.JSON(orders)
}
```

---

## ARCH-008: Prohibited Packages in Domain Layer
- **Code**: `LOY-ARCH-008`
- **Severity**: `ERROR`
- **Rule**: Domain files in `internal/.../domain` must never import external framework or driver packages:
  - `net/http`
  - `database/sql`
  - `github.com/gofiber/fiber`
  - `github.com/gin-gonic/gin`
  - `github.com/labstack/echo`
  - `gorm.io/gorm`
  - `github.com/redis/go-redis`
  - `github.com/valkey-io/valkey-go`
  - `google.golang.org/grpc`
- **Rationale**: The Domain layer represents your core business model and enterprise logic. It must remain pure Go without coupling to web protocols, database libraries, or caching SDKs.

### Prohibited Pattern
```go title="internal/customer/domain/customer.go"
package domain

import (
    "database/sql" // VIOLATION: Domain entity importing SQL driver!
    "net/http"     // VIOLATION: Domain entity importing HTTP transport!
)

type Customer struct {
    ID    int64
    Name  sql.NullString   // Leaking SQL driver types into domain!
    State http.Header      // Leaking HTTP types into domain!
}
```

### Permitted Remediation
Use standard Go primitive types (`string`, `*string`, `int`, `time.Time`) in domain entities:

```go title="internal/customer/domain/customer.go (Permitted)"
package domain

import "time"

type Customer struct {
    ID        int64
    Name      *string   // Use standard pointer for nullable values
    CreatedAt time.Time // Standard time.Time
}
```

---

## ARCH-009: Forbidden Transport Frameworks in Application Layer
- **Code**: `LOY-ARCH-009`
- **Severity**: `ERROR`
- **Rule**: Application service files in `internal/.../service` must never import web transport frameworks (`fiber`, `gin`, `echo`, `chi`).
- **Rationale**: Application services orchestrate business use cases. If a service method accepts `*fiber.Ctx` or returns `gin.H`, that service cannot be invoked from a background Asynq worker, a gRPC microservice server, or a CLI command.

### Prohibited Pattern
```go title="internal/notification/service/service.go"
package service

import "github.com/gofiber/fiber/v2" // VIOLATION: Application importing Fiber!

type NotificationService struct{}

func (s *NotificationService) Send(c *fiber.Ctx) error { // VIOLATION!
    // Cannot be called by background workers or CLI tasks!
}
```

### Permitted Remediation
Use standard `context.Context` and plain Go structs:

```go title="internal/notification/service/service.go (Permitted)"
package service

import "context"

type NotificationService struct{}

// Reusable across HTTP, gRPC, CLI, and Asynq background workers!
func (s *NotificationService) Send(ctx context.Context, userID int64, message string) error {
    // Process notification...
    return nil
}
```

---

## ARCH-010: Strict 4-Layer Direction Matrix
- **Code**: `LOY-ARCH-010`
- **Severity**: `ERROR`
- **Rule**: All package imports must strictly follow the allowed Clean Architecture directional matrix:
  - `Transport` may import `Application`, `Domain`, `Platform`
  - `Application` may import `Domain`, `Platform`
  - `Infrastructure` may import `Domain`, `Platform`
  - `Domain` may import `Platform` (pure utilities only)
- **Rationale**: Acts as the global fallback matrix enforcer across all workspace packages, preventing backwards layer dependencies (e.g. Infrastructure importing Application, or Domain importing Infrastructure).

### Checking Your Architecture
Run `loy check` to verify import purity:
```bash
loy check
```

Or generate an automated self-healing report for external AI agents:
```bash
loy check --format agent
```
