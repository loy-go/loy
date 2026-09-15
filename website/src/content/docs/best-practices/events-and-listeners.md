---
title: "Best Practices: Domain Events & Listeners"
description: "How to decouple cross-context side-effects and asynchronous reactions in Loy."
---

Generated via: `loy make event <name>` and `loy make listener <name>`

Domain Events capture business occurrences in the past tense (`OrderCreated`, `CandidateHired`, `InvoicePaid`). Listeners react to these events without coupling bounded contexts together.

## Golden Rules

### 1. Events Belong to Domain
- Declare event structs in `internal/<context>/domain/` or `internal/<context>/event/`:

```go title="internal/order/domain/event.go"
package domain

import "time"

// OrderCreatedEvent is emitted when a customer successfully completes checkout.
type OrderCreatedEvent struct {
	OrderID    int64
	CustomerID int64
	TotalCents int64
	OccurredAt time.Time
}
```

### 2. Side-Effects Belong in Listeners
- **DO NOT** trigger notification emails, analytics tracking, or external webhook calls directly inside your primary transaction service.
- **DO** publish a Domain Event and handle side-effects in independent listeners:

```go title="internal/notification/listener/order_created_listener.go"
package listener

import (
	"context"
	"fmt"
	"myapp/internal/order/domain"
)

type OrderCreatedListener struct {
	emailClient EmailClient
}

func (l *OrderCreatedListener) Handle(ctx context.Context, event domain.OrderCreatedEvent) error {
	return l.emailClient.SendOrderConfirmation(ctx, event.CustomerID, event.OrderID)
}
```

### 3. Reliable Dispatch via Transactional Outbox
- For events that cannot be lost if the process restarts, use the [Transactional Outbox Pattern](./transactional-outbox/) (`loy make outbox`) to commit the event inside the database transaction before broadcasting to external listeners.
