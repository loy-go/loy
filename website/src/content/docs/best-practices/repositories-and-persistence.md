---
title: "Best Practices: Repositories & Persistence Adapters"
description: "How to design consumer-owned repository interfaces and sqlc persistence adapters in Loy."
---

Generated via: `loy make repo <name>` or `loy make crud <name>`

Repositories provide an abstraction boundary between application use cases and the underlying persistence layer (PostgreSQL, SQLite, MySQL) ([ADR-015](/loy/adrs/)).

## Golden Rules

### 1. Consumer-Owned Interfaces
- **DO NOT** define repository interfaces in the database adapter package.
- **DO** declare repository interfaces in the consuming domain or application package:

```go title="internal/candidate/domain/repository.go"
package domain

import "context"

// Repository specifies persistence operations required by the candidate bounded context.
type Repository interface {
	FindByID(ctx context.Context, id int64) (*Candidate, error)
	FindAll(ctx context.Context, limit, offset int) ([]*Candidate, error)
	Create(ctx context.Context, entity *Candidate) error
	Update(ctx context.Context, entity *Candidate) error
	Delete(ctx context.Context, id int64) error
}
```

### 2. Concrete Implementation in Infrastructure
- Implement the interface in `internal/<context>/repository/` backed by `sqlc` or `pgx`:

```go title="internal/candidate/repository/postgres_repository.go"
package repository

import (
	"context"
	"database/sql"
	"fmt"
	"myapp/internal/candidate/domain"
)

type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(db *sql.DB) (*PostgresRepository, error) {
	if db == nil {
		return nil, fmt.Errorf("database connection pool is required")
	}
	return &PostgresRepository{db: db}, nil
}
```

### 3. PostgreSQL RLS Session Injection
- When multi-tenancy is active ([ADR-018](/loy/adrs/)), inject `SET LOCAL app.current_tenant_id = $1` inside a database transaction before query execution:

```go
func (r *PostgresRepository) WithTenantTx(ctx context.Context, tenantID string, fn func(tx *sql.Tx) error) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, "SET LOCAL app.current_tenant_id = $1", tenantID); err != nil {
		return fmt.Errorf("setting tenant session context: %w", err)
	}

	if err := fn(tx); err != nil {
		return err
	}

	return tx.Commit()
}
```

### 4. Zero Business Logic in SQL Adapters
- SQL queries and repository methods must only map database rows to domain entities. Validation, policy checks, and workflow decisions belong in Application Services.
