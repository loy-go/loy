---
title: "ARCH-011–015: Governance, State & Platform Purity"
description: "Authoritative specification for rules preventing global mutable state, reflection service locators, cross-application leakage, and platform impurity."
---

Beyond directional layer boundaries, Loy enforces core engineering hygiene rules to guarantee test isolation, prevent data races, and maintain workspace independence.

Rules **`ARCH-011` through `ARCH-015`** are evaluated during both Phase 1 AST inspection and Phase 2 deep type analysis by `loy check`.

---

## ARCH-011: Forbidden Reflection Service Locators

- **Diagnostic Code**: `LOY-ARCH-011`
- **Severity**: `ERROR`
- **Rule**: Prohibits reflection-based service locators, dynamic container lookups, and runtime dependency injection libraries (`uber/dig`, `sarulabs/di`, `facebookgo/inject`).
- **Rationale**: Reflection DI hides dependency requirements until runtime, causing applications to compile successfully but crash during production boot. Loy mandates explicit Go constructor injection in `internal/app/wiring.go` per [ADR-003](/loy/adrs/).

### Prohibited Pattern
```go title="internal/order/service/service.go (Prohibited)"
package service

import "github.com/example/di"

func NewOrderService() *Service {
    // VIOLATION [ARCH-011]: Runtime service locator lookup!
    repo := di.Get("orderRepository").(OrderRepository)
    return &Service{repo: repo}
}
```

### Compliant Pattern
Pass dependencies explicitly via constructor parameters:

```go title="internal/order/service/service.go (Compliant)"
package service

import "fmt"

type Service struct {
    repo OrderRepository
}

func NewService(repo OrderRepository) (*Service, error) {
    if repo == nil {
        return nil, fmt.Errorf("order repository is required")
    }
    return &Service{repo: repo}, nil
}
```

---

## ARCH-012: Forbidden Package-Level Mutable State

- **Diagnostic Code**: `LOY-ARCH-012`
- **Severity**: `ERROR`
- **Rule**: Package-level mutable variables (`var db *sql.DB`, `var currentTenant string`, `var globalConfig Config`) are strictly prohibited.
- **Rationale**: Global state causes subtle data race conditions under concurrent goroutines, makes parallel unit testing (`t.Parallel()`) impossible due to test cross-contamination, and introduces hidden coupling between packages.

### Prohibited Pattern
```go title="internal/user/repository/pg.go (Prohibited)"
package repository

import "database/sql"

// VIOLATION [ARCH-012]: Package-level mutable variable!
var GlobalDB *sql.DB

func FindUser(id int64) (*User, error) {
    return GlobalDB.QueryRow(...)
}
```

### Compliant Pattern
Encapsulate mutable state within struct instances created via explicit constructors:

```go title="internal/user/repository/pg.go (Compliant)"
package repository

import (
    "database/sql"
    "fmt"
)

type PostgresRepository struct {
    db *sql.DB
}

func NewPostgresRepository(db *sql.DB) (*PostgresRepository, error) {
    if db == nil {
        return nil, fmt.Errorf("database handle is required")
    }
    return &PostgresRepository{db: db}, nil
}
```

> **Allowed Exception**: Package-level constants (`const MaxRetries = 3`) and immutable sentinel errors (`var ErrNotFound = errors.New("not found")`) are permitted.

---

## ARCH-013: Cross-Application Monorepo Leakage

- **Diagnostic Code**: `LOY-ARCH-013`
- **Severity**: `ERROR`
- **Rule**: In multi-application monorepos (`apps/api` and `apps/worker`), one application daemon must never import internal code directly from another application daemon.
- **Rationale**: Direct cross-app imports create tight coupling between independent deployments, forcing synchronized releases and destroying modular boundary guarantees.

### Prohibited Pattern
```go title="apps/worker/internal/handler/job.go (Prohibited)"
package handler

// VIOLATION [ARCH-013]: Worker daemon imports internal API package!
import "github.com/example/monorepo/apps/api/internal/service"
```

### Compliant Pattern
Extract shared domain models or persistence adapters into a shared package under `packages/` or root `internal/`:

```go title="apps/worker/internal/handler/job.go (Compliant)"
package handler

import "github.com/example/monorepo/packages/shared/service"
```

---

## ARCH-014: Unmanaged Code in Generated Comment Regions

- **Diagnostic Code**: `LOY-ARCH-014`
- **Severity**: `WARN`
- **Rule**: Detects corrupt, nested, or unclosed comment regions (`// loy:region:...` / `// loy:endregion`) in managed files like `internal/app/wiring.go`.
- **Rationale**: Corrupt comment regions prevent atomic generators from splicing code safely and may cause syntax errors during automated scaffolding.

### Compliant Comment Region Anatomy
```go title="internal/app/wiring.go"
// loy:region:services
bookService, err := bookService.NewService(bookRepo)
if err != nil {
    return err
}
// loy:endregion
```

---

## ARCH-015: Platform Layer Purity

- **Diagnostic Code**: `LOY-ARCH-015`
- **Severity**: `ERROR`
- **Rule**: Domain and Application layers must never import impure platform utilities located in `internal/platform/*` or `pkg/*` ([ADR-023](/loy/adrs/)).
- **Definition of Impure Platform**: Any package that executes external I/O, binds to operating system handles, or imports database drivers (`net/http`, `database/sql`, `os`, `syscall`, `pgx`).

### Prohibited Pattern
```go title="internal/order/domain/order.go (Prohibited)"
package domain

// VIOLATION [ARCH-015]: Domain imports impure platform package!
import "github.com/example/bookstore/internal/platform/logger"

func (o *Order) Validate() error {
    logger.Info("validating order") // Impure I/O inside domain!
    return nil
}
```

### Compliant Pattern
Domain business logic must remain pure in-memory Go code. Log use-case events at the Application service layer, or emit Domain Events that are captured by Infrastructure listeners:

```go title="internal/order/service/service.go (Compliant)"
package service

import (
    "context"
    "log/slog"
)

func (s *Service) PlaceOrder(ctx context.Context, order *domain.Order) error {
    if err := order.Validate(); err != nil {
        return err
    }
    slog.InfoContext(ctx, "order placed successfully", "order_id", order.ID)
    return s.repo.Save(ctx, order)
}
```

---

## 6. Suppressing Rules with Explicit Rationale

When a valid edge case requires bypassing a rule, use explicit suppression annotations:

```go
// loy:ignore ARCH-012 reason="sync.Once initialized singleton for legacy third-party SDK"
var legacySDKClient *ThirdPartyClient
```

> **Mandatory Rule**: Every `// loy:ignore` comment **must** include an explicit `reason="..."` attribute. Suppressions missing a reason trigger a fatal error under `loy check`.
