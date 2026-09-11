package builtin

import (
	"context"
	"fmt"

	"github.com/uloydev/loy/internal/generator"
	"github.com/uloydev/loy/internal/generator/model"
)

// RuntimeData holds template values for runtime application scaffolding.
type RuntimeData struct {
	ModulePath    string
	HTTPFramework string // "fiber" or "nethttp"
}

// RuntimeGenerator scaffolds standard composition root, lifecycle coordinator, config and platform tools.
type RuntimeGenerator struct {
	modulePath    string
	httpFramework string
}

// NewRuntimeGenerator creates a RuntimeGenerator.
func NewRuntimeGenerator(modulePath, httpFramework string) *RuntimeGenerator {
	if httpFramework == "" {
		httpFramework = "fiber"
	}
	return &RuntimeGenerator{
		modulePath:    modulePath,
		httpFramework: httpFramework,
	}
}

func (g *RuntimeGenerator) Name() string {
	return "runtime"
}

func (g *RuntimeGenerator) Description() string {
	return "Scaffolds standard application runtime lifecycle, composition root, and health checks"
}

func (g *RuntimeGenerator) Generate(ctx context.Context, input generator.Input) ([]model.Artifact, error) {
	data := RuntimeData{
		ModulePath:    g.modulePath,
		HTTPFramework: g.httpFramework,
	}

	renderer := GetRenderer()

	// 1. cmd/api/main.go
	mainTmpl, err := ReadTemplate("runtime_main.go.tmpl")
	if err != nil {
		return nil, fmt.Errorf("reading runtime_main template: %w", err)
	}
	mainContent, err := renderer.RenderGo(ctx, "runtime_main", mainTmpl, data)
	if err != nil {
		return nil, fmt.Errorf("rendering main.go: %w", err)
	}

	// 2. internal/config/config.go
	configTmpl, err := ReadTemplate("runtime_config.go.tmpl")
	if err != nil {
		return nil, fmt.Errorf("reading runtime_config template: %w", err)
	}
	configContent, err := renderer.RenderGo(ctx, "runtime_config", configTmpl, data)
	if err != nil {
		return nil, fmt.Errorf("rendering config.go: %w", err)
	}

	// 3. internal/platform/logger/logger.go
	loggerTmpl, err := ReadTemplate("runtime_logger.go.tmpl")
	if err != nil {
		return nil, fmt.Errorf("reading runtime_logger template: %w", err)
	}
	loggerContent, err := renderer.RenderGo(ctx, "runtime_logger", loggerTmpl, data)
	if err != nil {
		return nil, fmt.Errorf("rendering logger.go: %w", err)
	}

	// 4. internal/platform/health/health.go
	healthTmpl, err := ReadTemplate("runtime_health.go.tmpl")
	if err != nil {
		return nil, fmt.Errorf("reading runtime_health template: %w", err)
	}
	healthContent, err := renderer.RenderGo(ctx, "runtime_health", healthTmpl, data)
	if err != nil {
		return nil, fmt.Errorf("rendering health.go: %w", err)
	}

	// 5. internal/platform/shutdown/coordinator.go
	shutdownTmpl, err := ReadTemplate("runtime_shutdown.go.tmpl")
	if err != nil {
		return nil, fmt.Errorf("reading runtime_shutdown template: %w", err)
	}
	shutdownContent, err := renderer.RenderGo(ctx, "runtime_shutdown", shutdownTmpl, data)
	if err != nil {
		return nil, fmt.Errorf("rendering coordinator.go: %w", err)
	}

	// 6. internal/app/app.go
	appTmpl, err := ReadTemplate("runtime_app.go.tmpl")
	if err != nil {
		return nil, fmt.Errorf("reading runtime_app template: %w", err)
	}
	appContent, err := renderer.RenderGo(ctx, "runtime_app", appTmpl, data)
	if err != nil {
		return nil, fmt.Errorf("rendering app.go: %w", err)
	}

	// 7. internal/app/wiring.go
	wiringTmpl, err := ReadTemplate("runtime_wiring.go.tmpl")
	if err != nil {
		return nil, fmt.Errorf("reading runtime_wiring template: %w", err)
	}
	wiringContent, err := renderer.RenderGo(ctx, "runtime_wiring", wiringTmpl, data)
	if err != nil {
		return nil, fmt.Errorf("rendering wiring.go: %w", err)
	}

	return []model.Artifact{
		{
			Path:        "cmd/api/main.go",
			Ownership:   model.DeveloperOwned,
			Permissions: 0644,
			Content:     mainContent,
		},
		{
			Path:        "internal/config/config.go",
			Ownership:   model.DeveloperOwned,
			Permissions: 0644,
			Content:     configContent,
		},
		{
			Path:        "internal/platform/logger/logger.go",
			Ownership:   model.DeveloperOwned,
			Permissions: 0644,
			Content:     loggerContent,
		},
		{
			Path:        "internal/platform/health/health.go",
			Ownership:   model.DeveloperOwned,
			Permissions: 0644,
			Content:     healthContent,
		},
		{
			Path:        "internal/platform/shutdown/coordinator.go",
			Ownership:   model.DeveloperOwned,
			Permissions: 0644,
			Content:     shutdownContent,
		},
		{
			Path:        "internal/app/app.go",
			Ownership:   model.DeveloperOwned,
			Permissions: 0644,
			Content:     appContent,
		},
		{
			Path:        "internal/app/wiring.go",
			Ownership:   model.MixedOwned,
			Permissions: 0644,
			Content:     wiringContent,
		},
	}, nil
}
