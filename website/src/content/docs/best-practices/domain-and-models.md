---
title: "Best Practices: Domain Models & Entities"
description: "How to design pure, robust domain entities and business models in Loy."
---

Generated via: `loy make model <name> [fields...]`

Domain models represent the heart of your business software. Under Loy's Clean Architecture rules ([Spec 05](/loy/rules/)), the **Domain Layer** must remain 100% pure Go.

## Golden Rules

### 1. Absolute Domain Purity
- **DO NOT** import `database/sql`, `pgx`, `gorm.io`, `net/http`, or web framework libraries into the Domain layer. This triggers rule `ARCH-008` or `ARCH-002` during `loy check`.
- **DO NOT** use database struct tags (like `gorm:"primaryKey"`) in domain models. Domain entities model business invariants, not table layouts.
- **DO** use standard Go types (`string`, `int64`, `time.Time`, `uuid.UUID`) and pure custom value objects.

```go title="internal/order/domain/order.go"
package domain

import (
	"errors"
	"time"
)

var (
	ErrEmptyOrderItems   = errors.New("order must contain at least one item")
	ErrInvalidOrderTotal = errors.New("order total must be greater than zero")
)

// Order represents the core business entity for customer orders.
type Order struct {
	ID         int64
	CustomerID int64
	TotalCents int64
	Status     OrderStatus
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// OrderStatus defines typed enumerated order states.
type OrderStatus string

const (
	StatusPending   OrderStatus = "pending"
	StatusPaid      OrderStatus = "paid"
	StatusCancelled OrderStatus = "cancelled"
)

// MarkPaid transitions the order to paid state ensuring state transition rules.
func (o *Order) MarkPaid() error {
	if o.Status == StatusCancelled {
		return errors.New("cannot pay for a cancelled order")
	}
	o.Status = StatusPaid
	o.UpdatedAt = time.Now().UTC()
	return nil
}
```

### 2. Rich Domain Models Over Anemic Data Bags
- Encapsulate mutation logic and validation methods directly on domain entity receiver methods (`order.Cancel()`, `candidate.AdvanceStage()`).
- Avoid exposing struct fields directly if business rules require state synchronization (e.g. updating `UpdatedAt` or recalculating total amounts).

### 3. Factory Constructors for Entity Creation
- Provide a `New<Entity>(...) (*<Entity>, error)` constructor inside the domain package to enforce creation invariants before the entity reaches services or repositories.
