package event

import (
	"time"
)

// UserRegisteredCreatedEvent triggers when a UserRegistered entity is created.
type UserRegisteredCreatedEvent struct {
	ID        int64     `json:"id"`
	Timestamp time.Time `json:"timestamp"`
}

// EventName returns the unique event identifier.
func (UserRegisteredCreatedEvent) EventName() string {
	return "user_registered.created"
}
