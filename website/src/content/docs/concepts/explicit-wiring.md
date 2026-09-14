---
title: "Explicit Constructor Wiring"
description: "Why Loy chooses explicit Go constructors in internal/app/wiring.go over runtime reflection DI."
---

In the broader Go ecosystem, framework abstractions often introduce heavy reflection dependency injection (e.g. `uber-go/dig`, `facebookgo/inject`) or compile-time codegen locators (`google/wire`).

Loy rejects runtime reflection DI and service locators in favor of **standard, explicit Go constructors** ([ADR-003](/loy/adrs/)).

## The Composition Root

Every Loy application defines a single composition root in `internal/app/wiring.go`:

```go title="internal/app/wiring.go"
package app

import (
	// loy:region:imports
	bookRepo "myapp/internal/book/repository"
	bookService "myapp/internal/book/service"
	bookHttp "myapp/internal/book/transport/http"
	// loy:endregion
)

func (a *App) wireDependencies() error {
	// loy:region:repositories
	bookRepo, err := bookRepo.NewPostgresRepository(a.db)
	if err != nil {
		return err
	}
	// loy:endregion

	// loy:region:services
	bookSvc, err := bookService.NewService(bookRepo)
	if err != nil {
		return err
	}
	// loy:endregion

	// loy:region:handlers
	bookHandler, err := bookHttp.NewHandler(bookSvc)
	if err != nil {
		return err
	}
	// loy:endregion

	// loy:region:routes
	bookHandler.RegisterRoutes(a.router.Group("/api/v1"))
	// loy:endregion

	return nil
}
```

---

## Why Explicit Wiring?

1. **Deterministic Compile-Time Safety**: If a constructor parameter changes, the standard Go compiler (`go build`) fails immediately with an accurate line number.
2. **Zero Runtime Reflection Overhead**: Startup time is instantaneous with zero memory allocations spent on type inspection.
3. **IDE Navigation & Debugging**: Jump-to-definition (`F12`), call hierarchies, and standard Go debuggers (`delve`) work naturally.
4. **Transparent Ownership**: You can trace every instantiated service and repository in one plain text file.
