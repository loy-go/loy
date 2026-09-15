---
title: "Best Practices: Composition Root & Modular Wiring"
description: "How to organize dependency injection and maintain small, readable wiring files across multiple bounded contexts."
---

Generated via: `internal/app/wiring.go` and `internal/app/wire_<domain>.go` ([ADR-003](/loy/adrs/), [ADR-020](/loy/adrs/))

The **Composition Root** is the single location in your application where concrete components are instantiated and wired together.

## Golden Rules

### 1. Explicit Go Constructors Only
- **DO NOT** use reflection DI libraries (`uber/dig`, `sarulabs/di`) or global service locators (`container.Get(...)`). This triggers `ARCH-011`.
- **DO** use explicit constructor injection (`NewPostgresRepository`, `NewService`, `NewHandler`). Missing dependencies fail at compile time.

### 2. Modular Sub-Domain Wire Files
- As your application grows beyond 5 bounded contexts, use the `--modular` flag (`loy make crud <name> --modular`) to split wiring into domain-specific files:

```
internal/app/
├── wiring.go            # Orchestrates sub-domain wire calls (< 40 LOC)
├── wire_iam.go          # IAM domain composition root
├── wire_billing.go      # Billing domain composition root
└── wire_candidate.go    # Candidate domain composition root
```

Each modular wire file defines a receiver method on `*App`:

```go title="internal/app/wire_candidate.go"
package app

import (
	candidateRepo "myapp/internal/candidate/repository"
	candidateService "myapp/internal/candidate/service"
	candidateHttp "myapp/internal/candidate/transport/http"
)

func (a *App) wireCandidate() error {
	repo, err := candidateRepo.NewPostgresRepository(a.db)
	if err != nil {
		return err
	}
	svc, err := candidateService.NewService(repo)
	if err != nil {
		return err
	}
	h, err := candidateHttp.NewHandler(svc)
	if err != nil {
		return err
	}
	if a.router != nil {
		h.RegisterRoutes(a.router.Group("/api/v1"))
	}
	return nil
}
```
