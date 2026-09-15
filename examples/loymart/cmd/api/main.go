package main

import (
	"context"
	"fmt"
	"os"

	"loymart/internal/app"
	"loymart/internal/config"
	"loymart/internal/platform/logger"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	log := logger.New(cfg.LogLevel, cfg.Environment)

	application, err := app.New(cfg, log)
	if err != nil {
		log.Error("failed to initialize application", "error", err)
		os.Exit(1)
	}

	ctx := context.Background()
	if err := application.Run(ctx); err != nil {
		log.Error("application terminated with error", "error", err)
		os.Exit(1)
	}
}
