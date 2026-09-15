---
title: "Best Practices: Application Services & Use Cases"
description: "How to orchestrate business workflows, transactions, and event dispatches in Loy services."
---

Generated via: `loy make service <name>` or `loy make crud <name>`

Application Services orchestrate workflows, validate permissions, execute business transactions, and dispatch domain events ([Spec 05](/loy/rules/)).

## Golden Rules

### 1. Transport-Agnostic Design
- **DO NOT** import `*fiber.Ctx`, `http.Request`, `http.ResponseWriter`, or Protobuf messages into Application Services. This triggers `ARCH-004` or `ARCH-009`.
- **DO** accept plain Go types (`context.Context`, `int64`, domain structs, or service DTOs) and return plain Go results or errors.

```go title="internal/candidate/service/service.go"
package service

import (
	"context"
	"fmt"
	"myapp/internal/candidate/domain"
)

type Service struct {
	repo domain.Repository
}

func NewService(repo domain.Repository) (*Service, error) {
	if repo == nil {
		return nil, fmt.Errorf("candidate repository dependency is required")
	}
	return &Service{repo: repo}, nil
}

func (s *Service) ScreenCandidate(ctx context.Context, id int64) (*domain.Candidate, error) {
	candidate, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("finding candidate %d: %w", id, err)
	}

	// Execute domain invariant state transition
	if err := candidate.MarkScreened(); err != nil {
		return nil, err
	}

	if err := s.repo.Update(ctx, candidate); err != nil {
		return nil, fmt.Errorf("updating candidate: %w", err)
	}

	return candidate, nil
}
```

### 2. Transaction Boundaries
- When a use case mutates multiple repositories or requires atomic outbox event persistence ([ADR-007](/loy/adrs/)), manage the transaction boundary at the service level using explicit transaction coordinators or unit of work adapters.

### 3. Fail Fast on Missing Dependencies
- Always validate required dependencies inside constructor functions (`NewService(...)`). Return a descriptive error if any dependency is `nil`, preventing `nil` pointer panics at runtime.
