package app

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"

	"loymart/internal/config"
	"loymart/internal/platform/health"
	"loymart/internal/platform/shutdown"
)

// State represents the runtime lifecycle state of the application.
type State string

const (
	StateCreated      State = "Created"
	StateInitializing State = "Initializing"
	StateReady        State = "Ready"
	StateRunning      State = "Running"
	StateStopping     State = "Stopping"
	StateStopped      State = "Stopped"
)

// App is the root composition application component.
type App struct {
	cfg         *config.Config
	log         *slog.Logger
	state       State
	mu          sync.RWMutex
	coordinator *shutdown.Coordinator
	healthReg   *health.Registry

	router *fiber.App

	db *sql.DB
}

// New creates and initializes a new App instance.
func New(cfg *config.Config, log *slog.Logger) (*App, error) {
	app := &App{
		cfg:         cfg,
		log:         log,
		state:       StateCreated,
		coordinator: shutdown.NewCoordinator(),
		healthReg:   health.NewRegistry(),
	}

	if err := app.init(); err != nil {
		// Clean up any resources created before failure
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = app.coordinator.Teardown(ctx)
		return nil, fmt.Errorf("initializing app: %w", err)
	}

	return app, nil
}

func (a *App) setState(s State) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.state = s
}

// State returns the current lifecycle state.
func (a *App) State() State {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.state
}

func (a *App) init() error {
	a.setState(StateInitializing)

	a.router = fiber.New(fiber.Config{
		DisableStartupMessage: true,
		ReadTimeout:           10 * time.Second,
		WriteTimeout:          15 * time.Second,
		IdleTimeout:           120 * time.Second,
		BodyLimit:             4 * 1024 * 1024,
	})

	// Register health probes
	a.router.Get("/health/live", adaptor.HTTPHandlerFunc(a.healthReg.LiveHandler()))
	a.router.Get("/health/ready", adaptor.HTTPHandlerFunc(a.healthReg.ReadyHandler()))

	// Wire application dependencies
	if err := a.wireDependencies(); err != nil {
		return fmt.Errorf("wiring dependencies: %w", err)
	}

	// Register transport shutdown hook

	a.coordinator.Register(shutdown.PhaseTransport, func(ctx context.Context) error {
		return a.router.ShutdownWithContext(ctx)
	})

	a.setState(StateReady)
	return nil
}

// Run starts the application and blocks until context cancellation or OS termination signal.
func (a *App) Run(ctx context.Context) error {
	a.setState(StateRunning)

	sigCtx, stopSignals := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stopSignals()

	serverErr := make(chan error, 1)

	go func() {

		addr := fmt.Sprintf(":%d", a.cfg.Port)
		a.log.Info("http server listening", "addr", addr, "framework", "fiber")
		if err := a.router.Listen(addr); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}

	}()

	shutdownTimeout := a.cfg.ShutdownTimeout
	if shutdownTimeout <= 0 {
		shutdownTimeout = 15 * time.Second
	}

	select {
	case <-sigCtx.Done():
		stopSignals()
		a.log.Info("shutdown signal received")
	case err := <-serverErr:
		shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()
		_ = a.Stop(shutdownCtx)
		return fmt.Errorf("server error: %w", err)
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	return a.Stop(shutdownCtx)
}

// Stop executes graceful shutdown across all phases.
func (a *App) Stop(ctx context.Context) error {
	a.setState(StateStopping)
	a.log.Info("starting graceful shutdown")

	err := a.coordinator.Teardown(ctx)
	a.setState(StateStopped)
	if err != nil {
		a.log.Error("graceful shutdown completed with errors", "error", err)
		return err
	}

	a.log.Info("graceful shutdown completed successfully")
	return nil
}
