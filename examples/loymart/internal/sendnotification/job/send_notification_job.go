package job

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hibiken/asynq"
)

// TypeSendNotificationProcess represents the Asynq task type.
const TypeSendNotificationProcess = "send_notification:process"

// SendNotificationPayload holds the background job parameters.
type SendNotificationPayload struct {
	ID int64 `json:"id"`
}

// NewSendNotificationTask creates a new Asynq task.
func NewSendNotificationTask(id int64) (*asynq.Task, error) {
	payload, err := json.Marshal(SendNotificationPayload{ID: id})
	if err != nil {
		return nil, fmt.Errorf("marshaling task payload: %w", err)
	}
	return asynq.NewTask(TypeSendNotificationProcess, payload), nil
}

// SendNotificationProcessor processes SendNotification background jobs.
type SendNotificationProcessor struct{}

// NewSendNotificationProcessor constructs a new task processor.
func NewSendNotificationProcessor() *SendNotificationProcessor {
	return &SendNotificationProcessor{}
}

// ProcessTask handles execution of the task.
func (p *SendNotificationProcessor) ProcessTask(ctx context.Context, t *asynq.Task) error {
	var payload SendNotificationPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("unmarshaling task payload: %w", err)
	}
	// Task processing logic
	return nil
}
