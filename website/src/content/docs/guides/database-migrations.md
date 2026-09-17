---
title: "Database Migrations & Safe DDL"
description: "Production guide to embedded Goose database migrations, zero-downtime DDL recipes, and PgBouncer connection pooling."
---

Database migrations in production require extreme care: an unindexed foreign key check, an unsafe `ALTER TABLE`, or a slow `CREATE INDEX` can take an `AccessExclusiveLock`, halting writes across your entire application.

Loy integrates an embedded **Goose migration engine** ([ADR-015](/loy/adrs/)) and provides specialized **zero-downtime migration recipes** designed for high-traffic PostgreSQL systems.

---

## 1. Zero-Downtime Migration Recipes

Loy scaffolds migration files with pre-configured safety preambles:

```sql
SET lock_timeout = '2s';
SET statement_timeout = '5s';
```

If a migration cannot acquire a table lock within 2 seconds, PostgreSQL immediately aborts the migration rather than queueing behind long-running queries and blocking all incoming application traffic.

### 1.1 Recipe: `index-concurrent` (Non-Blocking Indexes)
Standard `CREATE INDEX` locks the target table against all writes until index construction completes. On a table with millions of rows, this causes massive outages.

Scaffold a concurrent index migration:

```bash
loy make migration add_users_email_idx --recipe=index-concurrent --table=users
```

Generated SQL (`migrations/<timestamp>_add_users_email_idx.sql`):

```sql
-- +goose NO TRANSACTION
-- +goose Up
SET lock_timeout = '2s';
SET statement_timeout = '0s'; -- Allow background index build
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_users_email ON users (email);

-- +goose Down
DROP INDEX CONCURRENTLY IF EXISTS idx_users_email;
```

> **Key Rule**: Concurrent index builds cannot run inside a transaction block (`-- +goose NO TRANSACTION`).

---

### 1.2 Recipe: `shadow-column` (Zero-Downtime Type Changes)
Changing the type of a column or renaming a column directly locks the table and forces a full table rewrite. Loy provides a safe 3-phase shadow column recipe:

```bash
loy make migration change_user_age_to_bigint --recipe=shadow-column --table=users --type=BIGINT
```

Generated SQL:

```sql
-- +goose Up
SET lock_timeout = '2s';
SET statement_timeout = '5s';

-- 1. Add shadow column (instant metadata-only operation)
ALTER TABLE users ADD COLUMN IF NOT EXISTS age_shadow BIGINT;

-- 2. Create dual-write trigger function
CREATE OR REPLACE FUNCTION sync_users_age_shadow()
RETURNS TRIGGER AS $$
BEGIN
    NEW.age_shadow = NEW.age;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- 3. Attach trigger to capture concurrent writes
CREATE TRIGGER trg_sync_users_age_shadow
BEFORE INSERT OR UPDATE ON users
FOR EACH ROW EXECUTE FUNCTION sync_users_age_shadow();

-- +goose Down
DROP TRIGGER IF EXISTS trg_sync_users_age_shadow ON users;
DROP FUNCTION IF EXISTS sync_users_age_shadow();
ALTER TABLE users DROP COLUMN IF EXISTS age_shadow;
```

**Cutover Steps in Production**:
1. Run `loy migrate up` to create the shadow column and dual-write trigger.
2. Backfill existing historical rows in small batches:
   ```sql
   UPDATE users SET age_shadow = age WHERE age_shadow IS NULL AND id BETWEEN 1 AND 10000;
   ```
3. Deploy application code pointing to `age_shadow`.
4. Drop old trigger and original column.

---

### 1.3 Recipe: `raw` (Standard Migrations)
For standard table creations or additive column insertions:

```bash
loy make migration create_orders_table --recipe=raw
```

Generated SQL:

```sql
-- +goose Up
SET lock_timeout = '2s';
SET statement_timeout = '5s';

CREATE TABLE IF NOT EXISTS orders (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    public_id UUID NOT NULL DEFAULT gen_random_uuid(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp()
);

-- +goose Down
DROP TABLE IF EXISTS orders;
```

---

## 2. Tri-Route PostgreSQL Connection Pooling

In high-concurrency cloud environments using PgBouncer, prepared statement caching and transaction pooling can conflict. Loy scaffolds a **Tri-Route DB Pool** pattern in `internal/platform/database/postgres.go`:

```text
┌────────────────────────────────────────────────────────────────────────┐
│ 1. DataPool (*pgxpool.Pool)                                            │
│    • Configured with exec_mode_simple_protocol                         │
│    • Safe for PgBouncer transaction pooling mode                       │
│    • Handles 99% of application queries and repository calls           │
└────────────────────────────────────────────────────────────────────────┘

┌────────────────────────────────────────────────────────────────────────┐
│ 2. SessionPool (*pgxpool.Pool)                                         │
│    • Standard session pooling mode                                     │
│    • Supports prepared statements, complex batching, and migrations    │
└────────────────────────────────────────────────────────────────────────┘

┌────────────────────────────────────────────────────────────────────────┐
│ 3. DirectConn Factory                                                  │
│    • Direct unpooled *pgx.Conn connection                              │
│    • Dedicated for LISTEN / NOTIFY and PostgreSQL advisory locks       │
└────────────────────────────────────────────────────────────────────────┘
```

---

## 3. Migration CLI Operations

Loy embeds the Goose migration runner directly into the binary. No separate CLI tool is required:

### 3.1 Applying Migrations (`up`)
Run all pending database migrations:

```bash
loy migrate up
```

Apply a single next migration:

```bash
loy migrate up-by-one
```

---

### 3.2 Rolling Back Migrations (`down` & `reset`)
Roll back the most recently applied migration:

```bash
loy migrate down
```

Roll back all migrations to clean database state:

```bash
loy migrate reset
```

Redo (rollback then re-apply) the latest migration:

```bash
loy migrate redo
```

---

### 3.3 Inspecting Status
Inspect current migration history and pending status:

```bash
loy migrate status
```

Output:
```text
Migration Status:
  [Applied] 20260916000000_create_users_table.sql (applied at 2026-09-16 10:14:22 UTC)
  [Applied] 20260916000100_create_books_table.sql (applied at 2026-09-16 10:14:23 UTC)
  [Pending] 20260917000000_add_users_email_idx.sql
```
