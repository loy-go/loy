---
title: "Recipe: Zero-Docker Local Apps with SQLite"
description: "How to build lightweight CLI utilities, desktop tools, and local prototypes with zero Docker dependencies."
---

Not every Go application needs a heavyweight PostgreSQL or Redis daemon. For CLI tools, embedded desktop apps (Wails), and rapid local prototyping, Loy supports pure-Go **SQLite** persistence with zero CGO dependencies ([ADR-005](/loy/adrs/)).

## 1. Initializing SQLite Applications

Scaffold a project using Chi HTTP router and pure-Go SQLite:

```bash
loy new micro-tool --db=sqlite --http=chi --queue=inmemory
cd micro-tool
```

Generated `loy.yaml`:

```yaml title="loy.yaml"
version: 1
project:
  name: micro-tool

defaults:
  http: chi
  database: sqlite
  queue: inmemory
  cache: memory
```

## 2. Running Embedded Migrations

Embedded Goose automatically uses the SQLite SQL dialect. Scaffolding entities generates SQLite-compatible DDL:

```bash
loy make crud task "title:string:required,completed:bool"
```

Generated migration:

```sql title="migrations/20260914_create_tasks_table.sql"
-- +goose Up
CREATE TABLE IF NOT EXISTS tasks (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT NOT NULL,
    completed INTEGER NOT NULL DEFAULT 0,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- +goose Down
DROP TABLE IF EXISTS tasks;
```

## 3. Pure-Go Driver (`modernc.org/sqlite`)

Loy configures `modernc.org/sqlite`, meaning your Go application compiles with `CGO_ENABLED=0` for instantaneous cross-compilation to Linux, macOS, and Windows.

```bash
# Instant static binary build
CGO_ENABLED=0 go build -o bin/micro-tool ./cmd/api
```
