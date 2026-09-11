package queue

import (
	"github.com/hibiken/asynq"
)

// AsynqConfig configures redis/valkey connection and concurrency for Asynq.
type AsynqConfig struct {
	RedisAddr   string
	Password    string
	Concurrency int
}

// NewAsynqClient creates an Asynq client for enqueuing tasks.
func NewAsynqClient(cfg AsynqConfig) *asynq.Client {
	redisOpt := asynq.RedisClientOpt{
		Addr:     cfg.RedisAddr,
		Password: cfg.Password,
	}
	return asynq.NewClient(redisOpt)
}

// NewAsynqServer creates an Asynq background worker server.
func NewAsynqServer(cfg AsynqConfig) *asynq.Server {
	redisOpt := asynq.RedisClientOpt{
		Addr:     cfg.RedisAddr,
		Password: cfg.Password,
	}
	concurrency := cfg.Concurrency
	if concurrency <= 0 {
		concurrency = 10
	}
	return asynq.NewServer(redisOpt, asynq.Config{
		Concurrency: concurrency,
	})
}

// NewAsynqMux creates a new Asynq ServeMux router for registering task handlers.
func NewAsynqMux() *asynq.ServeMux {
	return asynq.NewServeMux()
}
