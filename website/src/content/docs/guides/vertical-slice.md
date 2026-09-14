---
title: "Building Vertical CRUD Slices"
description: "How to scaffold complete, clean-architecture features in seconds with loy make crud."
---

Vertical slicing organizes code around business features rather than technical layers. The `loy make crud` command generates an entire end-to-end slice in a single operation.

## Syntax & Field Descriptors

```bash
loy make crud <Name> [field:type[:modifier]...] [flags]
```

### Supported Field Types
- `string`: Go `string`, SQL `TEXT` or `VARCHAR`
- `int`, `int64`: Go integer, SQL `INTEGER` or `BIGINT`
- `float`, `float64`: Go floating point, SQL `NUMERIC` or `DOUBLE PRECISION`
- `bool`: Go boolean, SQL `BOOLEAN`
- `time`, `datetime`: Go `time.Time`, SQL `TIMESTAMPTZ`
- `uuid`: Go `uuid.UUID`, SQL `UUID`

### Field Modifiers
- `:required`: Disallows null or empty values.
- `:optional`: Generates pointer field (`*string`, `*int`) in Go struct.
- `:unique`: Adds database unique index constraint.
- `:index`: Adds database b-tree index.

---

## Example: Building a Customer Order Slice

```bash
loy make crud Order customer_id:int:required amount:float:required status:string:index
```

### Generated Artifacts
1. **Database Migration** (`migrations/*_create_orders_table.sql`): Timestamped goose DDL with Up and Down statements.
2. **SQL Query** (`queries/orders.sql`): Typed queries for SQLC (`GetOrderByID`, `ListOrders`, `CreateOrder`, `DeleteOrder`).
3. **Domain Entity** (`internal/order/domain/order.go`): Typed struct with JSON and DB tags.
4. **Repository Interface** (`internal/order/domain/repository.go`): Consumer-owned database interface.
5. **Postgres Adapter** (`internal/order/repository/pg_adapter.go`): Concrete implementation backed by PostgreSQL.
6. **Service Layer** (`internal/order/service/service.go`): Business use-case coordinator.
7. **HTTP Handler** (`internal/order/transport/http/handler.go`): Fiber/nethttp router endpoints.
8. **Unit Test** (`internal/order/service/service_test.go`): Mock repository test with testify.
9. **App Wiring** (`internal/app/wiring.go`): Spliced dependencies and registered routes.
