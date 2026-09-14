---
title: "Database Migrations & SQLC"
description: "Managing database schemas using Loy's embedded Goose migration engine and SQLC queries."
---

Loy embeds the battle-tested [Goose](https://github.com/pressly/goose) migration engine directly into the CLI binary ([ADR-015](/loy/adrs/)). You do not need to install an external migration binary.

## Migration Commands

### Create a New Migration

```bash
loy migrate create add_status_to_users
```

This creates a new timestamped file in `migrations/`:

```sql title="migrations/20260914120000_add_status_to_users.sql"
-- +goose Up
-- +goose StatementBegin
ALTER TABLE users ADD COLUMN status VARCHAR(50) NOT NULL DEFAULT 'active';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE users DROP COLUMN status;
-- +goose StatementEnd
```

---

### Apply Migrations (`up`)

```bash
# Using environment variable
export DATABASE_URL="postgres://postgres:secret@localhost:5432/myapp?sslmode=disable"
loy migrate up

# Or using CLI flag
loy migrate up --db-url="postgres://user:pass@localhost:5432/dbname"
```

---

### Roll Back Migrations (`down`)

```bash
# Roll back the single most recent migration
loy migrate down

# Roll back to a specific migration version
loy migrate down --to=20260914000000
```

---

### Inspect Migration Status

```bash
loy migrate status
```

---

## Type-Safe Queries with SQLC

Loy pairs Goose migrations with [SQLC](https://sqlc.dev/). Queries written in `queries/*.sql` are compiled into type-safe, reflection-free Go methods.

Whenever you add or modify a SQL query, run:

```bash
sqlc generate
```
