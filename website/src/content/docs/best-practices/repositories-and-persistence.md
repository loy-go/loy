---
title: "Best Practices: Repositories & Persistence Adapters"
description: "How to design consumer-owned repository interfaces, sqlc persistence adapters, transaction coordinators, and unit test doubles in Loy."
---

In Clean Architecture, the **Repository Pattern** decouples your business domain and use cases from the underlying database engine (PostgreSQL, SQLite, MySQL).

Scaffolded via:
```bash
# Scaffold an atomic repository interface and database adapter
loy make repo <name>

# Or scaffold a complete vertical CRUD slice with migrations and queries
loy make crud <name> [fields...]
```

---

## 1. Consumer-Owned Interfaces

A fundamental senior engineering rule in Go: **accept interfaces, return structs**, and **interfaces belong to the consumer, not the implementer**.

- **Incorrect**: Defining the `Repository` interface inside the database adapter package (`repository/postgres`).
- **Correct**: Declare the interface where it is consumed—in the `domain/` package:

```go title="internal/candidate/domain/repository.go"
package domain

import "context"

// Repository specifies persistence operations required by Candidate use cases.
type Repository interface {
	FindByID(ctx context.Context, id int64) (*Candidate, error)
	FindAll(ctx context.Context, limit, offset int) ([]*Candidate, error)
	Create(ctx context.Context, entity *Candidate) error
	Update(ctx context.Context, entity *Candidate) error
	Delete(ctx context.Context, id int64) error
}
```

By owning the interface in the Domain layer:
1. The domain dictates what storage behavior it needs.
2. The domain remains completely unaware of PostgreSQL, SQLite, or SQL syntax.
3. You can test application services instantly by supplying a mock in-memory struct.

---

## 2. Concrete Implementation in Infrastructure

Implement the interface in `internal/<feature>/repository/` backed by native database connection pools (`*sql.DB` or `*pgxpool.Pool`):

```go title="internal/candidate/repository/postgres_repository.go"
package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"myapp/internal/candidate/domain"
)

type PostgresRepository struct {
	db *sql.DB
}

// Explicit constructor injection (ADR-003)
func NewPostgresRepository(db *sql.DB) (*PostgresRepository, error) {
	if db == nil {
		return nil, fmt.Errorf("database connection pool is required")
	}
	return &PostgresRepository{db: db}, nil
}

func (r *PostgresRepository) FindByID(ctx context.Context, id int64) (*domain.Candidate, error) {
	const query = `
		SELECT id, name, email, status, created_at, updated_at
		FROM candidates
		WHERE id = $1
	`
	var c domain.Candidate
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&c.ID, &c.Name, &c.Email, &c.Status, &c.CreatedAt, &c.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrCandidateNotFound
		}
		return nil, fmt.Errorf("querying candidate %d: %w", id, err)
	}
	return &c, nil
}
```

---

## 3. Database Error Mapping

Never leak raw driver error strings (`pq: duplicate key value violates unique constraint`) out of the repository. Map raw database errors to typed **Domain Sentinel Errors**:

```go
// Map database driver codes to domain errors
if err != nil {
    if errors.Is(err, sql.ErrNoRows) {
        return domain.ErrCandidateNotFound
    }
    // PostgreSQL unique constraint error code (23505)
    var pgErr *pgconn.PgError
    if errors.As(err, &pgErr) && pgErr.Code == "23505" {
        return domain.ErrEmailAlreadyInUse
    }
    return fmt.Errorf("inserting candidate: %w", err)
}
```

This guarantees that Application services and Transport handlers can inspect errors with standard `errors.Is(err, domain.ErrEmailAlreadyInUse)` without importing database driver packages.

---

## 4. Multi-Tenant PostgreSQL Row-Level Security (RLS)

When multi-tenancy is active ([ADR-018](/loy/adrs/)), inject `SET LOCAL app.current_tenant_id = $1` inside a database transaction before query execution:

```go
func (r *PostgresRepository) WithTenantTx(ctx context.Context, tenantID string, fn func(tx *sql.Tx) error) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("beginning tenant transaction: %w", err)
	}
	defer tx.Rollback()

	// Sets RLS tenant context for this transaction only
	if _, err := tx.ExecContext(ctx, "SET LOCAL app.current_tenant_id = $1", tenantID); err != nil {
		return fmt.Errorf("setting tenant session context: %w", err)
	}

	if err := fn(tx); err != nil {
		return err
	}

	return tx.Commit()
}
```

---

## 5. Zero Business Logic in Repositories

Repositories are data mappers, **not decision makers**:
- **Prohibited in Repositories**: Checking whether an account has sufficient balance, calculating interest rates, sending email notifications, or verifying permissions.
- **Allowed in Repositories**: Executing SQL statements, scanning rows into structs, managing database transactions, and mapping database errors to domain errors.

---

## 6. Lightning-Fast Unit Testing with In-Memory Doubles

Because Application Services depend on `domain.Repository` rather than concrete PostgreSQL structs, you can write unit tests that run in microseconds without Docker or database setup:

```go title="internal/candidate/service/service_test.go"
package service_test

import (
	"context"
	"testing"
	"myapp/internal/candidate/domain"
	"myapp/internal/candidate/service"
)

// In-memory test double implementing domain.Repository
type mockCandidateRepo struct {
	candidates map[int64]*domain.Candidate
}

func (m *mockCandidateRepo) FindByID(ctx context.Context, id int64) (*domain.Candidate, error) {
	c, ok := m.candidates[id]
	if !ok {
		return nil, domain.ErrCandidateNotFound
	}
	return c, nil
}
func (m *mockCandidateRepo) Create(ctx context.Context, c *domain.Candidate) error {
	m.candidates[c.ID] = c
	return nil
}
// Implement remaining methods...

func TestCandidateService_GetCandidate(t *testing.T) {
	mockRepo := &mockCandidateRepo{
		candidates: map[int64]*domain.Candidate{
			1: {ID: 1, Name: "Alice"},
		},
	}

	svc, err := service.NewService(mockRepo)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	candidate, err := svc.GetCandidate(context.Background(), 1)
	if err != nil {
		t.Fatalf("failed to fetch candidate: %v", err)
	}

	if candidate.Name != "Alice" {
		t.Errorf("expected Alice, got %s", candidate.Name)
	}
}
```

This test runs in **0.001 seconds**, never touches a network socket, and tests business logic in total isolation.
