---
title: "Best Practices: Transactional Outbox Pattern"
description: "How to guarantee at-least-once message delivery using PostgreSQL SKIP LOCKED and Asynq in Loy."
---

Generated via: `loy make outbox`

The Transactional Outbox pattern guarantees that database state updates and event publishing to external message brokers occur atomically.

## Golden Rules

### 1. Atomic Persistence in Same Transaction
- Always insert the outbox record using the exact same `*sql.Tx` instance that writes the entity:

```go
func (r *Repository) CreateWithOutbox(ctx context.Context, tx *sql.Tx, entity *domain.Invoice) error {
	// 1. Write invoice
	if err := r.queries.WithTx(tx).CreateInvoice(ctx, entity); err != nil {
		return err
	}

	// 2. Write outbox event in same commit
	outboxStore := outbox.NewStore()
	return outboxStore.Save(ctx, tx, "invoice:created", entity)
}
```

### 2. High-Performance Polling with `SKIP LOCKED`
- The dispatcher query must use `FOR UPDATE SKIP LOCKED` and match the index on `(status, id ASC)`:

```sql
SELECT id, event_type, payload
FROM outbox_events
WHERE status = 'pending'
ORDER BY id ASC
LIMIT 50
FOR UPDATE SKIP LOCKED;
```

This ensures multiple worker replicas can poll the outbox concurrently without blocking each other or processing duplicate events.
