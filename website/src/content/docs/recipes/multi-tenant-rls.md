---
title: "Recipe: Multi-Tenant SaaS with PostgreSQL RLS"
description: "How to enforce bulletproof organization data isolation using PostgreSQL Row Level Security."
---

Multi-tenant architectures require absolute tenant isolation. Instead of polluting every application query with `WHERE org_id = $1` (which risks catastrophic data leaks if a single developer forgets a filter), Loy supports **PostgreSQL Row Level Security (RLS)** as a first-class citizen ([ADR-018](/loy/adrs/)).

## 1. Architecture Flow

```
HTTP Request ──► Tenant Middleware ──► Application Service ──► Repository (Postgres)
(Bearer JWT)     (Extracts org_id)     (Passes context.Context) (SET LOCAL app.current_tenant_id)
                                                                            │
                                                                   PostgreSQL Database
                                                               (Enforces RLS Policy)
```

## 2. Project Initialization

Create a project configured for PostgreSQL and RLS multi-tenancy:

```bash
loy new saas-platform --preset=api --db=postgres --multi-tenant=rls
cd saas-platform
```

In `loy.yaml`, multi-tenancy is activated:

```yaml title="loy.yaml"
version: 1
project:
  name: saas-platform

multi_tenancy:
  enabled: true
  strategy: rls
  tenant_key: org_id
```

## 3. Scaffolding Tenant-Aware CRUD

Run `loy make crud` to scaffold an entity:

```bash
loy make crud document "title:string:required,content:text" --modular
```

Loy automatically produces a migration with RLS policies enabled:

```sql title="migrations/20260914_create_documents_table.sql"
-- +goose Up
CREATE TABLE IF NOT EXISTS documents (
    id BIGSERIAL PRIMARY KEY,
    org_id UUID NOT NULL,
    title VARCHAR(255) NOT NULL,
    content TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_documents_org_id ON documents (org_id);

ALTER TABLE documents ENABLE ROW LEVEL SECURITY;
ALTER TABLE documents FORCE ROW LEVEL SECURITY;

CREATE POLICY documents_tenant_isolation_policy ON documents
    FOR ALL
    TO public
    USING (org_id = NULLIF(current_setting('app.current_tenant_id', true), '')::UUID);

-- +goose Down
DROP TABLE IF EXISTS documents CASCADE;
```

## 4. Query Execution with Tenant Context

Because RLS filters rows at the database engine level, sqlc queries remain standard, clean SQL without repetitive `WHERE org_id = $1`:

```sql title="queries/documents.sql"
-- name: ListDocuments :many
SELECT * FROM documents
ORDER BY id DESC
LIMIT $1 OFFSET $2;
```

When executing queries inside a transaction, set the session context:

```go
func (r *PostgresRepository) List(ctx context.Context, tenantID string, limit, offset int) ([]*domain.Document, error) {
    tx, err := r.db.BeginTx(ctx, nil)
    if err != nil {
        return nil, err
    }
    defer tx.Rollback()

    // Inject tenant session variable for this transaction
    _, err = tx.ExecContext(ctx, "SET LOCAL app.current_tenant_id = $1", tenantID)
    if err != nil {
        return nil, err
    }

    // Execute standard sqlc queries safely
    // PostgreSQL automatically filters documents WHERE org_id = tenantID
    return r.queries.WithTx(tx).ListDocuments(ctx, queries.ListDocumentsParams{
        Limit:  int32(limit),
        Offset: int32(offset),
    })
}
```

## 5. Verification

Verify architecture boundaries:

```bash
loy check --strict
```
