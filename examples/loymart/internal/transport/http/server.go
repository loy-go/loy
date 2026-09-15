package http

import (
	"context"
	"log/slog"
	"runtime/debug"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/google/uuid"
)

// Server wraps the Fiber application instance.
type Server struct {
	App    *fiber.App
	Logger *slog.Logger
	Addr   string
}

// Config provides setup parameters for Fiber HTTP transport.
type Config struct {
	Addr         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	AppName      string
}

// NewServer initializes Fiber application with standard Loy middleware.
func NewServer(cfg Config, logger *slog.Logger) *Server {
	if logger == nil {
		logger = slog.Default()
	}

	app := fiber.New(fiber.Config{
		AppName:      cfg.AppName,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			logger.Error("HTTP request failed",
				slog.String("path", c.Path()),
				slog.Int("status", code),
				slog.String("error", err.Error()),
			)
			return c.Status(code).JSON(fiber.Map{
				"error": err.Error(),
			})
		},
	})

	// 1. Panic Recovery Middleware
	app.Use(func(c *fiber.Ctx) (err error) {
		defer func() {
			if r := recover(); r != nil {
				stack := string(debug.Stack())
				logger.Error("panic recovered in HTTP handler",
					slog.Any("recover", r),
					slog.String("stack", stack),
					slog.String("path", c.Path()),
				)
				err = c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"error": "internal server error",
				})
			}
		}()
		return c.Next()
	})

	// 2. Request ID & Tracing Context Middleware
	app.Use(func(c *fiber.Ctx) error {
		reqID := c.Get("X-Request-ID")
		if reqID == "" {
			reqID = uuid.New().String()
		}
		c.Set("X-Request-ID", reqID)
		c.Locals("request_id", reqID)
		return c.Next()
	})

	// 3. Request Logging Middleware
	app.Use(func(c *fiber.Ctx) error {
		start := time.Now()
		err := c.Next()
		duration := time.Since(start)

		status := c.Response().StatusCode()
		reqID, _ := c.Locals("request_id").(string)

		logger.Info("HTTP request",
			slog.String("method", c.Method()),
			slog.String("path", c.Path()),
			slog.Int("status", status),
			slog.Duration("latency", duration),
			slog.String("request_id", reqID),
		)
		return err
	})

	// 4. Standard CORS Middleware
	app.Use(cors.New())

	return &Server{
		App:    app,
		Logger: logger,
		Addr:   cfg.Addr,
	}
}

// Start runs the HTTP server listener.
func (s *Server) Start() error {
	if s.Addr == "" {
		s.Addr = ":8080"
	}
	return s.App.Listen(s.Addr)
}

// Shutdown gracefully shuts down the Fiber application.
func (s *Server) Shutdown(ctx context.Context) error {
	return s.App.ShutdownWithContext(ctx)
}
