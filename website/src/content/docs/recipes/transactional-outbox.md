---
title: "Recipe: Transactional Outbox Pattern"
description: "How to guarantee zero-loss message delivery across database updates and message brokers."
---

When an application modifies a database row and publishes an event to a message broker (Asynq, Valkey, NATS, Kafka), a crash between the two writes causes inconsistent data state. The **Transactional Outbox Pattern** solves this by recording events inside the same database transaction.

## 1. Outbox Pipeline

```
HTTP Handler ──► SQL Transaction ──┬──► INSERT INTO orders (...)
                                   └──► INSERT INTO outbox_events (status='pending')
                                                              │
                                                cmd/worker Dispatcher
                                          (SELECT ... FOR UPDATE SKIP LOCKED)
                                                              │
                                                              ▼
                                                     Asynq / Valkey Broker
```

## 2. Scaffolding the Outbox Pipeline

Run `loy make outbox`:

```bash
loy make outbox
```

Scaffolded artifacts:
- **`migrations/<timestamp>_create_outbox_events_table.sql`**: Schema with indexed status column.
- **`internal/platform/outbox/store.go`**: Atomic transaction store helper.
- **`internal/platform/outbox/dispatcher.go`**: Background polling sweeper.

## 3. Atomic Event Persistence

In your repository, persist entity state and outbox events in the same transaction:

```go
func (r *OrderRepository) CreateOrder(ctx context.Context, order *domain.Order) error {
    tx, err := r.db.BeginTx(ctx, nil)
    if err != nil {
        return err
    }
    defer tx.Rollback()

    // 1. Insert order
    if err := r.queries.WithTx(tx).InsertOrder(ctx, order); err != nil {
        return err
    }

    // 2. Insert outbox record in same commit
    outboxStore := outbox.NewStore()
    if err := outboxStore.Save(ctx, tx, "order:created", order); err != nil {
        return err
    }

    return tx.Commit()
}
```

## 4. Polling & Dispatching Events

The outbox dispatcher runs in `cmd/worker`, safely polling events using `FOR UPDATE SKIP LOCKED` to prevent duplicate processing:

```go
dispatcher := outbox.NewDispatcher(db, asynqPublisher, 1*time.Second)
go dispatcher.Run(ctx)
```
