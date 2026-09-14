package builtin

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/loy-go/loy/internal/generator"
	"github.com/loy-go/loy/internal/generator/model"
)

// RuntimeData holds template values for runtime application scaffolding.
type RuntimeData struct {
	ModulePath    string
	ProjectName   string
	HTTPFramework string // "fiber" or "nethttp"
	WithDatabase  bool
	WithCache     bool
	WithQueue     bool
	WithTelemetry bool
	Template      string // e.g. "templ"
	Assets        string // e.g. "vite"
	Entrypoint    string // "api" or "web"
}

// RuntimeGenerator scaffolds standard composition root, lifecycle coordinator, config and platform tools.
type RuntimeGenerator struct {
	modulePath    string
	httpFramework string
	withDatabase  bool
	withCache     bool
	withQueue     bool
	withTelemetry bool
	template      string
	assets        string
	entrypoint    string
}

// NewRuntimeGenerator creates a RuntimeGenerator.
func NewRuntimeGenerator(modulePath, httpFramework string) *RuntimeGenerator {
	if httpFramework == "" {
		httpFramework = "fiber"
	}
	return &RuntimeGenerator{
		modulePath:    modulePath,
		httpFramework: httpFramework,
		withDatabase:  true,
		withCache:     true,
		withQueue:     true,
		withTelemetry: true,
		entrypoint:    "api",
	}
}

// WithCapabilities configures optional capability toggles.
func (g *RuntimeGenerator) WithCapabilities(db, cache, queue, telemetry bool) *RuntimeGenerator {
	g.withDatabase = db
	g.withCache = cache
	g.withQueue = queue
	g.withTelemetry = telemetry
	return g
}

// WithTemplate sets the template engine (e.g. "templ").
func (g *RuntimeGenerator) WithTemplate(t string) *RuntimeGenerator {
	g.template = t
	return g
}

// WithAssets sets the assets bundler (e.g. "vite").
func (g *RuntimeGenerator) WithAssets(a string) *RuntimeGenerator {
	g.assets = a
	return g
}

// WithEntrypoint sets the entrypoint type ("api" or "web").
func (g *RuntimeGenerator) WithEntrypoint(e string) *RuntimeGenerator {
	g.entrypoint = e
	return g
}

func (g *RuntimeGenerator) Name() string {
	return "runtime"
}

func (g *RuntimeGenerator) Description() string {
	return "Scaffolds standard application runtime lifecycle, composition root, and health checks"
}

func (g *RuntimeGenerator) Generate(ctx context.Context, input generator.Input) ([]model.Artifact, error) {
	projectName := filepath.Base(g.modulePath)
	if projectName == "." || projectName == "/" || projectName == "" {
		projectName = "myapp"
	}

	entrypoint := g.entrypoint
	if entrypoint == "" {
		entrypoint = "api"
	}

	data := RuntimeData{
		ModulePath:    g.modulePath,
		ProjectName:   projectName,
		HTTPFramework: g.httpFramework,
		WithDatabase:  g.withDatabase,
		WithCache:     g.withCache,
		WithQueue:     g.withQueue,
		WithTelemetry: g.withTelemetry,
		Template:      g.template,
		Assets:        g.assets,
		Entrypoint:    entrypoint,
	}

	renderer := GetRenderer()

	// 1. cmd/<entrypoint>/main.go
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

	mainPath := "cmd/api/main.go"
	if entrypoint == "web" {
		mainPath = "cmd/web/main.go"
	}

	artifacts := []model.Artifact{
		{
			Path:        mainPath,
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
	}

	// 8. Conditionally add platform integration templates
	if g.withDatabase {
		pgTmpl, err := ReadTemplate("platform_postgres.go.tmpl")
		if err != nil {
			return nil, fmt.Errorf("reading platform_postgres template: %w", err)
		}
		pgContent, err := renderer.RenderGo(ctx, "platform_postgres", pgTmpl, data)
		if err != nil {
			return nil, fmt.Errorf("rendering platform_postgres: %w", err)
		}
		artifacts = append(artifacts, model.Artifact{
			Path:        "internal/platform/database/postgres.go",
			Ownership:   model.DeveloperOwned,
			Permissions: 0644,
			Content:     pgContent,
		})
	}

	if g.withCache {
		valkeyTmpl, err := ReadTemplate("platform_valkey.go.tmpl")
		if err != nil {
			return nil, fmt.Errorf("reading platform_valkey template: %w", err)
		}
		valkeyContent, err := renderer.RenderGo(ctx, "platform_valkey", valkeyTmpl, data)
		if err != nil {
			return nil, fmt.Errorf("rendering platform_valkey: %w", err)
		}
		artifacts = append(artifacts, model.Artifact{
			Path:        "internal/platform/cache/valkey.go",
			Ownership:   model.DeveloperOwned,
			Permissions: 0644,
			Content:     valkeyContent,
		})
	}

	if g.withQueue {
		asynqTmpl, err := ReadTemplate("platform_asynq.go.tmpl")
		if err != nil {
			return nil, fmt.Errorf("reading platform_asynq template: %w", err)
		}
		asynqContent, err := renderer.RenderGo(ctx, "platform_asynq", asynqTmpl, data)
		if err != nil {
			return nil, fmt.Errorf("rendering platform_asynq: %w", err)
		}
		artifacts = append(artifacts, model.Artifact{
			Path:        "internal/platform/queue/asynq.go",
			Ownership:   model.DeveloperOwned,
			Permissions: 0644,
			Content:     asynqContent,
		})
	}

	if g.withTelemetry {
		otelTmpl, err := ReadTemplate("platform_otel.go.tmpl")
		if err != nil {
			return nil, fmt.Errorf("reading platform_otel template: %w", err)
		}
		otelContent, err := renderer.RenderGo(ctx, "platform_otel", otelTmpl, data)
		if err != nil {
			return nil, fmt.Errorf("rendering platform_otel: %w", err)
		}
		artifacts = append(artifacts, model.Artifact{
			Path:        "internal/platform/telemetry/otel.go",
			Ownership:   model.DeveloperOwned,
			Permissions: 0644,
			Content:     otelContent,
		})
	}

	if g.httpFramework == "fiber" {
		fiberTmpl, err := ReadTemplate("transport_fiber.go.tmpl")
		if err != nil {
			return nil, fmt.Errorf("reading transport_fiber template: %w", err)
		}
		fiberContent, err := renderer.RenderGo(ctx, "transport_fiber", fiberTmpl, data)
		if err != nil {
			return nil, fmt.Errorf("rendering transport_fiber: %w", err)
		}
		artifacts = append(artifacts, model.Artifact{
			Path:        "internal/transport/http/server.go",
			Ownership:   model.DeveloperOwned,
			Permissions: 0644,
			Content:     fiberContent,
		})
	}

	// 9. Fullstack / Web Templ integration
	if g.template == "templ" {
		renderTmpl, err := ReadTemplate("fullstack_render.go.tmpl")
		if err != nil {
			return nil, fmt.Errorf("reading fullstack_render template: %w", err)
		}
		renderContent, err := renderer.RenderGo(ctx, "fullstack_render", renderTmpl, data)
		if err != nil {
			return nil, fmt.Errorf("rendering view/render.go: %w", err)
		}
		artifacts = append(artifacts, model.Artifact{
			Path:        "internal/transport/http/view/render.go",
			Ownership:   model.DeveloperOwned,
			Permissions: 0644,
			Content:     renderContent,
		})

		webHandlerTmpl, err := ReadTemplate("fullstack_web_handler.go.tmpl")
		if err != nil {
			return nil, fmt.Errorf("reading fullstack_web_handler template: %w", err)
		}
		webHandlerContent, err := renderer.RenderGo(ctx, "fullstack_web_handler", webHandlerTmpl, data)
		if err != nil {
			return nil, fmt.Errorf("rendering handler/web.go: %w", err)
		}
		artifacts = append(artifacts, model.Artifact{
			Path:        "internal/transport/http/handler/web.go",
			Ownership:   model.DeveloperOwned,
			Permissions: 0644,
			Content:     webHandlerContent,
		})

		baseTmpl, err := ReadTemplate("fullstack_base.templ.tmpl")
		if err != nil {
			return nil, fmt.Errorf("reading fullstack_base template: %w", err)
		}
		baseContent, err := renderer.Render(ctx, "fullstack_base", baseTmpl, data)
		if err != nil {
			return nil, fmt.Errorf("rendering base.templ: %w", err)
		}
		artifacts = append(artifacts, model.Artifact{
			Path:        "views/layouts/base.templ",
			Ownership:   model.DeveloperOwned,
			Permissions: 0644,
			Content:     baseContent,
		})

		navbarTmpl, err := ReadTemplate("fullstack_navbar.templ.tmpl")
		if err != nil {
			return nil, fmt.Errorf("reading fullstack_navbar template: %w", err)
		}
		navbarContent, err := renderer.Render(ctx, "fullstack_navbar", navbarTmpl, data)
		if err != nil {
			return nil, fmt.Errorf("rendering navbar.templ: %w", err)
		}
		artifacts = append(artifacts, model.Artifact{
			Path:        "views/components/navbar.templ",
			Ownership:   model.DeveloperOwned,
			Permissions: 0644,
			Content:     navbarContent,
		})

		alertTmpl, err := ReadTemplate("fullstack_alert.templ.tmpl")
		if err != nil {
			return nil, fmt.Errorf("reading fullstack_alert template: %w", err)
		}
		alertContent, err := renderer.Render(ctx, "fullstack_alert", alertTmpl, data)
		if err != nil {
			return nil, fmt.Errorf("rendering alert.templ: %w", err)
		}
		artifacts = append(artifacts, model.Artifact{
			Path:        "views/components/alert.templ",
			Ownership:   model.DeveloperOwned,
			Permissions: 0644,
			Content:     alertContent,
		})

		homeTmpl, err := ReadTemplate("fullstack_home.templ.tmpl")
		if err != nil {
			return nil, fmt.Errorf("reading fullstack_home template: %w", err)
		}
		homeContent, err := renderer.Render(ctx, "fullstack_home", homeTmpl, data)
		if err != nil {
			return nil, fmt.Errorf("rendering home.templ: %w", err)
		}
		artifacts = append(artifacts, model.Artifact{
			Path:        "views/pages/home.templ",
			Ownership:   model.DeveloperOwned,
			Permissions: 0644,
			Content:     homeContent,
		})
	}

	// 10. Vite & frontend asset pipeline
	if g.assets == "vite" {
		viteTemplates := []struct {
			tmplName string
			outPath  string
			isGo     bool
		}{
			{"fullstack_package.json.tmpl", "package.json", false},
			{"fullstack_vite.config.ts.tmpl", "vite.config.ts", false},
			{"fullstack_tailwind.config.js.tmpl", "tailwind.config.js", false},
			{"fullstack_postcss.config.js.tmpl", "postcss.config.js", false},
			{"fullstack_app.css.tmpl", "assets/css/app.css", false},
			{"fullstack_app.ts.tmpl", "assets/js/app.ts", false},
			{"fullstack_assets.go.tmpl", "internal/transport/http/view/assets.go", true},
			{"fullstack_static.go.tmpl", "internal/transport/http/static/static.go", true},
		}

		for _, vt := range viteTemplates {
			rawTmpl, err := ReadTemplate(vt.tmplName)
			if err != nil {
				return nil, fmt.Errorf("reading %s: %w", vt.tmplName, err)
			}
			var rendered []byte
			if vt.isGo {
				rendered, err = renderer.RenderGo(ctx, vt.tmplName, rawTmpl, data)
			} else {
				rendered, err = renderer.Render(ctx, vt.tmplName, rawTmpl, data)
			}
			if err != nil {
				return nil, fmt.Errorf("rendering %s: %w", vt.outPath, err)
			}
			artifacts = append(artifacts, model.Artifact{
				Path:        vt.outPath,
				Ownership:   model.DeveloperOwned,
				Permissions: 0644,
				Content:     rendered,
			})
		}
	}

	return artifacts, nil
}
