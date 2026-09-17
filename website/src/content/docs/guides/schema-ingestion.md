---
title: "Schema-First Ingestion"
description: "Reverse-engineer and scaffold complete Clean Architecture slices from OpenAPI specs, JSON Schema, or PostgreSQL databases."
---

Modern engineering teams often work from existing contracts or databases:
1. **Contract-First Teams**: Product and API architects specify contracts in OpenAPI 3.x before backend engineers write a single line of code.
2. **Brownfield Migrations**: Teams migrating legacy monolithic databases (Rails, Django, Laravel, Node.js) into high-performance Go microservices need to reverse-engineer existing database tables without manual boilerplate transcription.

Loy supports both workflows out of the box via `loy make from-spec` and `loy make from-db`.

---

## 1. Ingestion from OpenAPI 3.x & JSON Schema

Scaffold production Clean Architecture slices directly from API specifications:

```bash
loy make from-spec ./specs/petstore.yaml
```

### 1.1 Specification Support
`loy make from-spec` parses both JSON and YAML specifications across:
- **OpenAPI 3.0 & 3.1**: Discovers models under `components.schemas.*`.
- **Swagger 2.0 & JSON Schema**: Discovers models under `definitions.*`.
- **Top-Level Schemas**: Single-entity JSON schema objects with top-level `properties`.

### 1.2 Type & Constraint Mapping
Loy extracts property definitions and translates them into typed Loy field specifications:

| OpenAPI Property Definition | Loy Field Modifier | Generated Go Type |
|---|---|---|
| `type: string` | `field:string` | `string` |
| `type: integer, format: int64` | `field:int64` | `int64` |
| `type: integer` | `field:int` | `int` |
| `type: number` | `field:float` | `float64` |
| `type: boolean` | `field:bool` | `bool` |
| `type: string, format: date-time` | `field:time` | `time.Time` |
| In `required` list | `:required` | Non-pointer with validation |
| Not in `required` list | *(optional)* | Pointer `*T` with `omitempty` |
| `enum: [active, pending, archived]` | `:enum[active,pending,archived]` | Custom enum type validation |

### 1.3 Generated Architecture Slice
For every schema entity discovered in the specification, Loy generates:
- Domain Entity with business validation methods (`internal/<domain>/domain/<entity>.go`)
- Consumer-owned Repository Interface (`internal/<domain>/domain/repository.go`)
- Concrete PostgreSQL Repository Adapter (`internal/<domain>/repository/pg_adapter.go`)
- Application Service Use Case (`internal/<domain>/service/service.go`)
- HTTP Handler with JSON DTO validation (`internal/<domain>/transport/http/handler.go`)
- SQLC Query file (`queries/<entities>.sql`)
- Database Migration file (`migrations/<timestamp>_create_<entities>_table.sql`)
- Explicit Dependency Wiring registration in `internal/app/wiring.go`

---

## 2. Ingestion from PostgreSQL Databases

Reverse-engineer existing PostgreSQL database schemas into Clean Architecture slices:

```bash
loy make from-db "postgres://user:password@localhost:5432/production_db?sslmode=disable" --tables=orders,order_items
```

### 2.1 Table Filtering
By default, Loy introspects all tables in the `public` schema. Use `--tables` to selectively ingest specific tables:

```bash
loy make from-db "postgres://..." --tables=users,invoices,accounts
```

### 2.2 Information Schema Introspection
Loy executes parameterized queries against `information_schema.columns`:
- Discovers column names, data types, and nullability (`is_nullable = 'NO'`).
- Automatically skips system and framework metadata columns (`id`, `public_id`, `created_at`, `updated_at`, `org_id`) to avoid collision with Loy's managed lifecycle fields.
- Maps PostgreSQL types (`bigint` -> `int64`, `numeric` -> `float`, `timestamptz` -> `time.Time`, `text`/`varchar` -> `string`).
- Automatically singularizes plural table names (`orders` -> `order`, `categories` -> `category`) following idiomatic Go domain naming rules.

---

## 3. Ingestion from SQL DDL Schema Files

When connecting to a live database is prohibited by network boundaries or security policies, Loy can parse offline SQL DDL files:

```bash
loy make from-db ./migrations/schema.sql
```

### Example DDL File:
```sql title="schema.sql"
CREATE TABLE IF NOT EXISTS products (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    public_id UUID NOT NULL DEFAULT gen_random_uuid(),
    sku VARCHAR(50) NOT NULL UNIQUE,
    name TEXT NOT NULL,
    price NUMERIC(10, 2) NOT NULL,
    stock INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp()
);
```

Running `loy make from-db ./schema.sql` automatically extracts:
- Entity: `product`
- Fields: `sku:string:required:unique name:string:required price:float:required stock:int:required`
- Immediately scaffolds the entire Clean Architecture vertical slice.
