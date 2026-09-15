package app

import (
	"github.com/hibiken/asynq"
	// loy:region:worker_imports
	sendnotificationJob "loymart/internal/sendnotification/job"
	// loy:endregion
)

// WireWorkerTasks registers all background job handlers with the Asynq ServeMux.
func (a *App) WireWorkerTasks(mux *asynq.ServeMux) error {
	// loy:region:tasks
	mux.HandleFunc(sendnotificationJob.TypeSendNotificationProcess, sendnotificationJob.NewSendNotificationProcessor().ProcessTask)
	// loy:endregion

	return nil
}
