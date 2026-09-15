package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/valkey-io/valkey-go"
)

// ValkeyConfig holds connection parameters for Valkey / Redis.
type ValkeyConfig struct {
	InitAddress []string
	Password    string
	SelectDB    int
}

// NewValkeyClient initializes a Valkey client pool and pings the server.
func NewValkeyClient(ctx context.Context, cfg ValkeyConfig) (valkey.Client, error) {
	if len(cfg.InitAddress) == 0 {
		cfg.InitAddress = []string{"127.0.0.1:6379"}
	}

	client, err := valkey.NewClient(valkey.ClientOption{
		InitAddress: cfg.InitAddress,
		Password:    cfg.Password,
		SelectDB:    cfg.SelectDB,
	})
	if err != nil {
		return nil, fmt.Errorf("creating valkey client: %w", err)
	}

	pingCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	cmd := client.B().Ping().Build()
	if err := client.Do(pingCtx, cmd).Error(); err != nil {
		client.Close()
		return nil, fmt.Errorf("pinging valkey server: %w", err)
	}

	return client, nil
}
