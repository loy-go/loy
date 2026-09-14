package builtin_test

import (
	"context"
	"strings"
	"testing"

	"github.com/loy-go/loy/internal/generator"
	"github.com/loy-go/loy/internal/generator/builtin"
)

func TestRuntimeGenerator_Fiber(t *testing.T) {
	gen := builtin.NewRuntimeGenerator("github.com/example/demo", "fiber")
	if gen.Name() != "runtime" {
		t.Fatalf("expected name 'runtime', got %q", gen.Name())
	}

	artifacts, err := gen.Generate(context.Background(), generator.Input{Name: "runtime"})
	if err != nil {
		t.Fatalf("generation failed: %v", err)
	}

	expectedPaths := map[string]bool{
		"cmd/api/main.go":                           false,
		"internal/config/config.go":                 false,
		"internal/platform/logger/logger.go":         false,
		"internal/platform/health/health.go":         false,
		"internal/platform/shutdown/coordinator.go": false,
		"internal/app/app.go":                       false,
		"internal/app/wiring.go":                    false,
	}

	for _, art := range artifacts {
		expectedPaths[art.Path] = true
	}

	for path, found := range expectedPaths {
		if !found {
			t.Errorf("missing expected artifact: %s", path)
		}
	}

	// Verify Fiber-specific bindings
	var appContent, mainContent, configContent, healthContent, shutdownContent string
	for _, art := range artifacts {
		switch art.Path {
		case "internal/app/app.go":
			appContent = string(art.Content)
		case "cmd/api/main.go":
			mainContent = string(art.Content)
		case "internal/config/config.go":
			configContent = string(art.Content)
		case "internal/platform/health/health.go":
			healthContent = string(art.Content)
		case "internal/platform/shutdown/coordinator.go":
			shutdownContent = string(art.Content)
		}
	}

	if !strings.Contains(appContent, "github.com/gofiber/fiber/v2") {
		t.Errorf("expected Fiber import in app.go")
	}
	if !strings.Contains(appContent, "StateCreated") || !strings.Contains(appContent, "StateRunning") {
		t.Errorf("expected 6 lifecycle states in app.go")
	}
	if !strings.Contains(mainContent, "app.New(cfg, log)") {
		t.Errorf("expected app.New in main.go")
	}
	if !strings.Contains(configContent, "ShutdownTimeout time.Duration") {
		t.Errorf("expected ShutdownTimeout in config.go")
	}
	if !strings.Contains(healthContent, "LiveHandler") || !strings.Contains(healthContent, "ReadyHandler") {
		t.Errorf("expected health handlers in health.go")
	}
	if !strings.Contains(shutdownContent, "PhaseTransport") || !strings.Contains(shutdownContent, "PhaseStorage") {
		t.Errorf("expected phase buckets in coordinator.go")
	}
}

func TestRuntimeGenerator_NetHTTP(t *testing.T) {
	gen := builtin.NewRuntimeGenerator("github.com/example/demo", "nethttp")

	artifacts, err := gen.Generate(context.Background(), generator.Input{Name: "runtime"})
	if err != nil {
		t.Fatalf("generation failed: %v", err)
	}

	var appContent string
	for _, art := range artifacts {
		if art.Path == "internal/app/app.go" {
			appContent = string(art.Content)
			break
		}
	}

	if strings.Contains(appContent, "github.com/gofiber/fiber") {
		t.Errorf("expected no Fiber in nethttp app.go")
	}
	if !strings.Contains(appContent, "http.ServeMux") {
		t.Errorf("expected http.ServeMux in nethttp app.go")
	}
}
