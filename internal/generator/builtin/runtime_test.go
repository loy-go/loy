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

	var appContent, serverContent string
	for _, art := range artifacts {
		if art.Path == "internal/app/app.go" {
			appContent = string(art.Content)
		}
		if art.Path == "internal/transport/http/server.go" {
			serverContent = string(art.Content)
		}
	}

	if strings.Contains(appContent, "github.com/gofiber/fiber") {
		t.Errorf("expected no Fiber in nethttp app.go")
	}
	if !strings.Contains(appContent, "http.ServeMux") {
		t.Errorf("expected http.ServeMux in nethttp app.go")
	}
	if !strings.Contains(serverContent, "http.NewServeMux()") {
		t.Errorf("expected http.NewServeMux in nethttp server.go")
	}
}

func TestRuntimeGenerator_Chi(t *testing.T) {
	gen := builtin.NewRuntimeGenerator("github.com/example/demo", "chi")

	artifacts, err := gen.Generate(context.Background(), generator.Input{Name: "runtime"})
	if err != nil {
		t.Fatalf("generation failed: %v", err)
	}

	var appContent, serverContent string
	for _, art := range artifacts {
		if art.Path == "internal/app/app.go" {
			appContent = string(art.Content)
		}
		if art.Path == "internal/transport/http/server.go" {
			serverContent = string(art.Content)
		}
	}

	if !strings.Contains(appContent, "github.com/go-chi/chi/v5") {
		t.Errorf("expected Chi import in app.go")
	}
	if !strings.Contains(appContent, "*chi.Mux") {
		t.Errorf("expected chi.Mux in app.go")
	}
	if !strings.Contains(serverContent, "chi.NewRouter()") {
		t.Errorf("expected chi.NewRouter() in chi server.go")
	}
}

func TestRuntimeGenerator_Gin(t *testing.T) {
	gen := builtin.NewRuntimeGenerator("github.com/example/demo", "gin")

	artifacts, err := gen.Generate(context.Background(), generator.Input{Name: "runtime"})
	if err != nil {
		t.Fatalf("generation failed: %v", err)
	}

	var appContent, serverContent string
	for _, art := range artifacts {
		if art.Path == "internal/app/app.go" {
			appContent = string(art.Content)
		}
		if art.Path == "internal/transport/http/server.go" {
			serverContent = string(art.Content)
		}
	}

	if !strings.Contains(appContent, "github.com/gin-gonic/gin") {
		t.Errorf("expected Gin import in app.go")
	}
	if !strings.Contains(appContent, "*gin.Engine") {
		t.Errorf("expected gin.Engine in app.go")
	}
	if !strings.Contains(serverContent, "gin.New()") {
		t.Errorf("expected gin.New() in gin server.go")
	}
}

func TestRuntimeGenerator_Sqlite(t *testing.T) {
	gen := builtin.NewRuntimeGenerator("github.com/example/demo", "fiber").
		WithDatabaseDialect("sqlite")

	artifacts, err := gen.Generate(context.Background(), generator.Input{Name: "runtime"})
	if err != nil {
		t.Fatalf("generation failed: %v", err)
	}

	var hasSqlite, hasPostgres bool
	var configContent string
	for _, art := range artifacts {
		if art.Path == "internal/platform/database/sqlite.go" {
			hasSqlite = true
			if !strings.Contains(string(art.Content), "modernc.org/sqlite") {
				t.Errorf("missing modernc.org/sqlite in sqlite.go")
			}
		}
		if art.Path == "internal/platform/database/postgres.go" {
			hasPostgres = true
		}
		if art.Path == "internal/config/config.go" {
			configContent = string(art.Content)
		}
	}

	if !hasSqlite {
		t.Errorf("expected internal/platform/database/sqlite.go")
	}
	if hasPostgres {
		t.Errorf("did not expect postgres.go when sqlite is chosen")
	}
	if !strings.Contains(configContent, `envDefault:"app.db"`) {
		t.Errorf("expected app.db default in config for sqlite")
	}
}

func TestRuntimeGenerator_MySQL(t *testing.T) {
	gen := builtin.NewRuntimeGenerator("github.com/example/demo", "fiber").
		WithDatabaseDialect("mysql")

	artifacts, err := gen.Generate(context.Background(), generator.Input{Name: "runtime"})
	if err != nil {
		t.Fatalf("generation failed: %v", err)
	}

	var hasMySQL bool
	var configContent string
	for _, art := range artifacts {
		if art.Path == "internal/platform/database/mysql.go" {
			hasMySQL = true
			if !strings.Contains(string(art.Content), "github.com/go-sql-driver/mysql") {
				t.Errorf("missing go-sql-driver/mysql in mysql.go")
			}
		}
		if art.Path == "internal/config/config.go" {
			configContent = string(art.Content)
		}
	}

	if !hasMySQL {
		t.Errorf("expected internal/platform/database/mysql.go")
	}
	if !strings.Contains(configContent, `envDefault:"root:root@tcp(localhost:3306)/app?parseTime=true"`) {
		t.Errorf("expected mysql default url in config")
	}
	if !strings.Contains(configContent, `strings.Contains(cfg.DatabaseURL, "root:root@")`) {
		t.Errorf("expected mysql root credential validation in config")
	}
}

func TestRuntimeGenerator_DatabaseNone(t *testing.T) {
	gen := builtin.NewRuntimeGenerator("github.com/example/demo", "fiber").
		WithDatabaseDialect("none")

	artifacts, err := gen.Generate(context.Background(), generator.Input{Name: "runtime"})
	if err != nil {
		t.Fatalf("generation failed: %v", err)
	}

	for _, art := range artifacts {
		if strings.HasPrefix(art.Path, "internal/platform/database/") {
			t.Errorf("did not expect database artifacts when database is 'none', got %s", art.Path)
		}
	}
}
