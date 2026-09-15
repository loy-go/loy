package listener

import (
	"context"
	"log/slog"

	"loymart/internal/userregistered/event"
)

// OnUserRegisteredCreatedListener handles UserRegisteredCreatedEvent.
type OnUserRegisteredCreatedListener struct {
	logger *slog.Logger
}

// NewOnUserRegisteredCreatedListener constructs a new listener.
func NewOnUserRegisteredCreatedListener(logger *slog.Logger) *OnUserRegisteredCreatedListener {
	return &OnUserRegisteredCreatedListener{logger: logger}
}

// Handle processes the received event.
func (l *OnUserRegisteredCreatedListener) Handle(ctx context.Context, evt event.UserRegisteredCreatedEvent) error {
	if l.logger != nil {
		l.logger.InfoContext(ctx, "processing event", "event", evt.EventName(), "id", evt.ID)
	}
	return nil
}
