---
title: "CQRS & Segregated Pipelines"
description: "Production guide to Command Query Responsibility Segregation (CQRS), segregated pipelines, and distributed idempotency deduplication with Loy."
---

Command Query Responsibility Segregation (CQRS) separates data modification operations (Commands) from data retrieval operations (Queries). While standard CRUD works well for straightforward administrative entities, high-throughput systems often require distinct optimization paths for writes versus reads.

Loy provides native scaffolding for CQRS without imposing any proprietary runtime frameworks or reflection-based event buses.

---

## 1. When to Use CRUD vs. CQRS

```text
┌──────────────────────────────────────┬──────────────────────────────────────┐
│ Standard CRUD (loy make crud)        │ CQRS Architecture (command / query)  │
├──────────────────────────────────────┼──────────────────────────────────────┤
│ Single entity model for read & write │ Segregated Command & Query pipelines │
│ 1:1 mapping between HTTP and Table   │ Task-based UI with intentional actions│
│ Simple state transitions             │ Complex business invariants & events │
│ Uniform read/write traffic ratio     │ Read traffic heavily outpaces writes │
│ Direct domain aggregate hydration    │ Direct SQLC reads bypass aggregates  │
└──────────────────────────────────────┴──────────────────────────────────────┘
```

---

## 2. Command Pipeline (State Mutations)

Commands represent intentional state-changing business tasks (e.g. `CheckoutCart`, `CancelSubscription`, `ApproveInvoice`).

### 2.1 Generating a Command

```bash
loy make command checkout_cart "cart_id:string total:float"
```

This scaffolds `internal/<domain>/command/checkout_cart_cmd.go`:

```go title="internal/cart/command/checkout_cart_cmd.go"
package command

import (
	"context"
	"fmt"
)

// CheckoutCartCommand encapsulates the input parameters required for the command.
type CheckoutCartCommand struct {
	CartId string
	Total  float64
}

// CheckoutCartHandler executes the transactional command.
type CheckoutCartHandler struct {
	// Inject database connection pool, outbox dispatcher, or domain repositories
}

func NewCheckoutCartHandler() *CheckoutCartHandler {
	return &CheckoutCartHandler{}
}

func (h *CheckoutCartHandler) Handle(ctx context.Context, cmd CheckoutCartCommand) error {
	// 1. Begin atomic database transaction
	// 2. Hydrate aggregate from repository
	// 3. Execute domain business validation: aggregate.Checkout()
	// 4. Save aggregate mutation
	// 5. Append domain event to transactional outbox table
	// 6. Commit transaction
	return nil
}
```

### 2.2 Transactional Outbox Coordination
Commands typically emit domain events upon completion. By writing the event directly into an `outbox` table within the **same database transaction** that mutates business entities, Loy guarantees at-least-once message delivery without distributed two-phase commits (2PC).

---

## 3. Query Pipeline (Read Optimization)

Queries represent read-only requests. In traditional DDD, loading a complex aggregate just to display three fields on a dashboard causes massive object allocation overhead and N+1 database queries.

In Loy, Queries **bypass the Domain Layer entirely**:
- Queries connect directly to the database or a read-replica pool.
- Queries execute specialized, projection-optimized SQL queries compiled via SQLC.
- Queries return formatted DTOs directly to the Transport layer with zero aggregate hydration.

```text
Query Flow:
  HTTP Handler ──> Query DTO ──> Query Handler ──> SQLC Query ──> Read Replica DB
                                                            │
                                                            ▼
                                                    Fast JSON Response
                                              (Zero Domain Aggregate Allocations!)
```

### 3.1 Generating a Query

```bash
loy make query order_summary "order_id:string total:float status:string"
```

This scaffolds `internal/<domain>/query/order_summary_query.go`:

```go title="internal/order/query/order_summary_query.go"
package query

import (
	"context"
	"fmt"
)

// OrderSummaryQuery represents the read request parameters.
type OrderSummaryQuery struct {
	OrderId string
	Total   float64
	Status  string
}

// OrderSummaryResult represents the projected read model.
type OrderSummaryResult struct {
	Data any
}

// OrderSummaryHandler executes the read operation.
type OrderSummaryHandler struct{}

func NewOrderSummaryHandler() *OrderSummaryHandler {
	return &OrderSummaryHandler{}
}

func (h *OrderSummaryHandler) Handle(ctx context.Context, q OrderSummaryQuery) (*OrderSummaryResult, error) {
	// Execute fast, optimized SQL projection query directly via sqlc
	return &OrderSummaryResult{}, nil
}
```

---

## 4. Distributed Idempotency Engine

Network retries, payment gateway webhooks, and impatient users clicking buttons twice can cause duplicate mutations. Loy provides a production-grade idempotency engine powered by PostgreSQL row locks.

### 4.1 Scaffolding Idempotency

```bash
loy make idempotency
```

This command generates two critical artifacts:

#### 1. Database Migration (`migrations/<timestamp>_create_idempotency_keys_table.sql`)
```sql
-- +goose Up
CREATE TABLE IF NOT EXISTS idempotency_keys (
    key VARCHAR(255) PRIMARY KEY,
    user_id VARCHAR(100) NOT NULL DEFAULT '',
    status VARCHAR(50) NOT NULL DEFAULT 'started',
    response_code INT NOT NULL DEFAULT 0,
    response_body TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp(),
    expires_at TIMESTAMPTZ NOT NULL DEFAULT (clock_timestamp() + INTERVAL '24 hours')
);

CREATE INDEX IF NOT EXISTS idx_idempotency_keys_expires_at ON idempotency_keys (expires_at);

-- +goose Down
DROP TABLE IF EXISTS idempotency_keys;
```

#### 2. HTTP Middleware (`internal/platform/middleware/idempotency.go`)
The middleware coordinates execution using PostgreSQL row-level locks:

```text
Incoming Request with "Idempotency-Key: abc-123"
     │
     ▼
Check idempotency_keys table
     ├─ Key exists and status == 'finished' ──> Return cached response immediately!
     ├─ Key exists and status == 'started'  ──> Conflict: Concurrent request in flight!
     └─ Key does not exist:
           │
           ▼
        Insert record with status 'started'
           │
           ▼
        Execute Command / Handler
           │
           ▼
        Update record with status 'finished' and cache response code & body
```

### 4.2 Handling Concurrent Replays Safely
When two identical requests arrive simultaneously with the same `Idempotency-Key`:
- The first request inserts the key with status `started`.
- The second request encounters a unique constraint collision or row-lock hold (`SELECT ... FOR UPDATE SKIP LOCKED`).
- The second request is safely rejected with HTTP 409 Conflict or held until the first finishes, preventing duplicate ledger entries or double-charges.
