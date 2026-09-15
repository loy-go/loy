---
title: "Best Practices: Database Migrations & Seeders"
description: "How to manage schema evolution, zero-downtime DDL, and reproducible development seeders in Loy."
---

Managed via: `loy migrate` and `loy seed` ([ADR-015](/loy/adrs/))

Loy embeds the mature `goose` migration engine directly into the application binary, eliminating external migration runner dependencies.

## Golden Rules

### 1. Always Write Reversible Migrations
- Every `-- +goose Up` migration block must have a corresponding, verified `-- +goose Down` rollback block.
- Verify rollback safety locally by running `loy migrate redo` before committing new migrations.

```sql title="migrations/20260914_create_candidates_table.sql"
-- +goose Up
CREATE TABLE IF NOT EXISTS candidates (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- +goose Down
DROP TABLE IF EXISTS candidates CASCADE;
```

### 2. Idempotent Development Seeders
- Scaffold seeders with `loy make seeder <name>`.
- Use `ON CONFLICT DO NOTHING` or check for existing records to ensure running `loy seed` multiple times is safe and deterministic:

```go title="internal/platform/database/seeds/candidate_seeder.go"
package seeds

import (
	"context"
	"database/sql"
)

type CandidateSeeder struct{}

func (s *CandidateSeeder) Seed(ctx context.Context, db *sql.DB) error {
	query := `
		INSERT INTO candidates (id, name, created_at, updated_at)
		VALUES (1, 'Alice Developer', NOW(), NOW())
		ON CONFLICT (id) DO NOTHING;
	`
	_, err := db.ExecContext(ctx, query)
	return err
}
```

### 3. Track Executed Seeds
- `loy seed` executes seeders inside database transactions and tracks execution history in the `schema_seeds` table.
