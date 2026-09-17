---
title: "Best Practices: Composition Root & Modular Wiring"
description: "Authoritative guide to explicit dependency injection, composition root architecture, and modular sub-domain wiring in Loy."
---

In Clean Architecture, individual packages (`domain`, `service`, `repository`, `transport`) accept narrow interfaces and never instantiate concrete dependencies.

The **Composition Root** is the single designated location in your application where concrete components are instantiated and wired together ([ADR-003](/loy/adrs/)). In Loy, this is located at `internal/app/wiring.go`.

---

## 1. Golden Rules of Dependency Wiring

### Rule 1: Explicit Go Constructors Only
- **Strictly Prohibited**: Reflection-based DI containers (`uber/dig`, `sarulabs/di`) or global service locators (`container.Get("userService")`). These trigger rule **`ARCH-011`**.
- **Mandated**: Standard Go constructors (`NewPostgresRepository(db)`, `NewService(repo)`). Missing or nil dependencies fail immediately at compile time or during fast startup validation.

### Rule 2: Single Direction of Initialization
Dependencies are always initialized from lowest infrastructure up to outer transport:

```text
Database Connection Pool (*pgxpool.Pool)
    │
    ▼
Repository Adapter (internal/order/repository)
    │
    ▼
Application Service (internal/order/service)
    │
    ▼
Transport Handler (internal/order/transport/http)
    │
    ▼
Router Registration (router.Group("/api/v1"))
```

---

## 2. Standard Monolithic Wiring (`internal/app/wiring.go`)

For small to medium projects (1–5 bounded contexts), all wiring is coordinated directly in `internal/app/wiring.go`:

```go title="internal/app/wiring.go"
package app

import (
	"fmt"

	// Feature package imports
	userRepo "github.com/example/bookstore/internal/user/repository"
	userSvc "github.com/example/bookstore/internal/user/service"
	userHttp "github.com/example/bookstore/internal/user/transport/http"

	orderRepo "github.com/example/bookstore/internal/order/repository"
	orderSvc "github.com/example/bookstore/internal/order/service"
	orderHttp "github.com/example/bookstore/internal/order/transport/http"
)

// wireDependencies coordinates explicit constructor injection for the API daemon.
func (a *App) wireDependencies() error {
	// =========================================================================
	// 1. Repositories
	// =========================================================================
	// loy:region:repositories
	uRepo, err := userRepo.NewPostgresRepository(a.db)
	if err != nil {
		return fmt.Errorf("wiring user repository: %w", err)
	}

	oRepo, err := orderRepo.NewPostgresRepository(a.db)
	if err != nil {
		return fmt.Errorf("wiring order repository: %w", err)
	}
	// loy:endregion

	// =========================================================================
	// 2. Services / Use Cases
	// =========================================================================
	// loy:region:services
	uService, err := userSvc.NewService(uRepo)
	if err != nil {
		return fmt.Errorf("wiring user service: %w", err)
	}

	oService, err := orderSvc.NewService(oRepo)
	if err != nil {
		return fmt.Errorf("wiring order service: %w", err)
	}
	// loy:endregion

	// =========================================================================
	// 3. Handlers & Router Registration
	// =========================================================================
	// loy:region:handlers
	uHandler, err := userHttp.NewHandler(uService)
	if err != nil {
		return fmt.Errorf("wiring user handler: %w", err)
	}

	oHandler, err := orderHttp.NewHandler(oService)
	if err != nil {
		return fmt.Errorf("wiring order handler: %w", err)
	}
	// loy:endregion

	// Mount routes onto HTTP router group
	if a.router != nil {
		apiGroup := a.router.Group("/api/v1")
		uHandler.RegisterRoutes(apiGroup)
		oHandler.RegisterRoutes(apiGroup)
	}

	return nil
}
```

---

## 3. Modular Sub-Domain Composition Roots (`--modular`)

As systems scale beyond 5 bounded contexts (e.g. `iam`, `billing`, `catalog`, `shipping`, `analytics`), a monolithic `wiring.go` can exceed 1,000 lines.

Loy resolves this via **Modular Sub-Domain Composition Roots** ([ADR-020](/loy/adrs/)):

```bash
loy make crud billing_account --modular
```

This decouples the composition root into domain-specific wire files:

```text
internal/app/
├── wiring.go            # Orchestrates sub-domain wire calls (< 40 lines)
├── wire_iam.go          # IAM domain composition root
├── wire_billing.go      # Billing domain composition root
└── wire_catalog.go      # Catalog domain composition root
```

### 3.1 Sub-Domain Wire File (`wire_billing.go`)
Each modular wire file defines a private receiver method on `*App`:

```go title="internal/app/wire_billing.go"
package app

import (
	"fmt"

	billingRepo "github.com/example/bookstore/internal/billing/repository"
	billingSvc "github.com/example/bookstore/internal/billing/service"
	billingHttp "github.com/example/bookstore/internal/billing/transport/http"
)

func (a *App) wireBilling() error {
	repo, err := billingRepo.NewPostgresRepository(a.db)
	if err != nil {
		return fmt.Errorf("wiring billing repository: %w", err)
	}

	svc, err := billingSvc.NewService(repo)
	if err != nil {
		return fmt.Errorf("wiring billing service: %w", err)
	}

	h, err := billingHttp.NewHandler(svc)
	if err != nil {
		return fmt.Errorf("wiring billing handler: %w", err)
	}

	if a.router != nil {
		h.RegisterRoutes(a.router.Group("/api/v1/billing"))
	}

	return nil
}
```

### 3.2 Main Wiring Orchestrator
The central `wiring.go` becomes a concise orchestrator:

```go title="internal/app/wiring.go"
func (a *App) wireDependencies() error {
	if err := a.wireIAM(); err != nil {
		return err
	}
	if err := a.wireBilling(); err != nil {
		return err
	}
	if err := a.wireCatalog(); err != nil {
		return err
	}
	return nil
}
```

---

## 4. Background Worker Wiring (`internal/app/worker_wiring.go`)

Loy applications operate under a **Dual-Daemon Topography** ([ADR-019](/loy/adrs/)):
- `cmd/api/main.go`: Dedicated HTTP and WebSocket server.
- `cmd/worker/main.go`: Dedicated Asynq background worker daemon.

The worker daemon has its own isolated composition root in `internal/app/worker_wiring.go`. This guarantees that background workers never initialize unnecessary HTTP routers or web middleware:

```go title="internal/app/worker_wiring.go"
package app

import (
	"github.com/hibiken/asynq"
)

func (w *WorkerApp) registerTaskHandlers(mux *asynq.ServeMux) error {
	// Register background job processors directly
	mux.HandleFunc("email:send_welcome", w.handleSendWelcomeEmail)
	mux.HandleFunc("report:generate_monthly", w.handleGenerateMonthlyReport)
	return nil
}
```

---

## 5. Unit Testing the Composition Root

Because `wiring.go` accepts typed structs rather than global variables, you can verify your entire application dependency graph in an automated unit test without connecting to external networks:

```go title="internal/app/wiring_test.go"
package app_test

import (
	"testing"
	"github.com/loy-go/loy/internal/filesystem"
)

func TestAppWiring(t *testing.T) {
	// Verify that dependencies initialize without nil pointer errors
	app, err := NewTestApp()
	if err != nil {
		t.Fatalf("failed to initialize application composition root: %v", err)
	}

	if app.Router() == nil {
		t.Errorf("expected router to be initialized")
	}
}
```
