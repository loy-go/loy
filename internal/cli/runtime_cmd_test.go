package cli_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/loy-go/loy/internal/cli"
	"github.com/loy-go/loy/internal/filesystem"
)

func TestNewProjectGeneratesRuntime(t *testing.T) {
	memFS := filesystem.NewMemFileSystem()
	rootCmd := cli.NewRootCmdWithFS(memFS, nil)

	rootCmd.SetArgs([]string{"new", "myapi", "--preset", "api"})
	err := rootCmd.ExecuteContext(context.Background())
	if err != nil {
		t.Fatalf("loy new failed: %v", err)
	}

	// CleanAndValidatePath resolves relative to current dir, so files are in abs path
	targetDir, _ := filesystem.CleanAndValidatePath(".", "myapi")

	expectedFiles := []string{
		targetDir + "/loy.yaml",
		targetDir + "/go.mod",
		targetDir + "/cmd/api/main.go",
		targetDir + "/internal/config/config.go",
		targetDir + "/internal/platform/logger/logger.go",
		targetDir + "/internal/platform/health/health.go",
		targetDir + "/internal/platform/shutdown/coordinator.go",
		targetDir + "/internal/app/app.go",
		targetDir + "/internal/app/wiring.go",
	}

	for _, file := range expectedFiles {
		exists, _ := memFS.Exists(file)
		if !exists {
			t.Errorf("expected file %s to be created by loy new", file)
		}
	}

	appBytes, _ := memFS.ReadFile(targetDir + "/internal/app/app.go")
	if !strings.Contains(string(appBytes), "github.com/gofiber/fiber/v2") {
		t.Errorf("expected fiber framework in generated api app.go")
	}
}

func TestMakeRuntimeCommand(t *testing.T) {
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(origDir) }()

	tmpDir := t.TempDir()
	memFS := filesystem.NewOSFileSystem()
	_ = memFS.MkdirAll(tmpDir, 0755)
	_ = memFS.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte("module github.com/test/workspace\n\ngo 1.22\n"), 0644)
	_ = memFS.WriteFile(filepath.Join(tmpDir, "loy.yaml"), []byte("version: 1\nproject:\n  name: workspace\n"), 0644)

	rootCmd := cli.NewRootCmdWithFS(memFS, nil)
	rootCmd.SetArgs([]string{"make", "runtime", "--http", "nethttp", "--db", "sqlite", "-C", tmpDir})
	err = rootCmd.ExecuteContext(context.Background())
	if err != nil {
		t.Fatalf("loy make runtime failed: %v", err)
	}

	appFile := filepath.Join(tmpDir, "internal", "app", "app.go")
	exists, _ := memFS.Exists(appFile)
	if !exists {
		t.Fatalf("expected internal/app/app.go to be generated")
	}

	content, _ := memFS.ReadFile(appFile)
	if !strings.Contains(string(content), "http.ServeMux") {
		t.Errorf("expected net/http server in app.go")
	}

	sqliteFile := filepath.Join(tmpDir, "internal", "platform", "database", "sqlite.go")
	if sqExists, _ := memFS.Exists(sqliteFile); !sqExists {
		t.Errorf("expected sqlite.go to be generated via make runtime --db sqlite")
	}

	// Test make runtime with --db none
	tmpDirNone := t.TempDir()
	_ = memFS.MkdirAll(tmpDirNone, 0755)
	_ = memFS.WriteFile(filepath.Join(tmpDirNone, "go.mod"), []byte("module github.com/test/nonedb\n\ngo 1.22\n"), 0644)
	_ = memFS.WriteFile(filepath.Join(tmpDirNone, "loy.yaml"), []byte("version: 1\nproject:\n  name: nonedb\n"), 0644)

	rootCmdNone := cli.NewRootCmdWithFS(memFS, nil)
	rootCmdNone.SetArgs([]string{"make", "runtime", "--http", "fiber", "--db", "none", "-C", tmpDirNone, "--force"})
	if err := rootCmdNone.ExecuteContext(context.Background()); err != nil {
		t.Fatalf("loy make runtime --db none failed: %v", err)
	}

	pgFile := filepath.Join(tmpDirNone, "internal", "platform", "database", "postgres.go")
	if pgExists, _ := memFS.Exists(pgFile); pgExists {
		t.Errorf("did not expect postgres.go when make runtime has --db none")
	}
}

func TestNewProjectGeneratesFullstackRuntime(t *testing.T) {
	memFS := filesystem.NewMemFileSystem()
	rootCmd := cli.NewRootCmdWithFS(memFS, nil)

	rootCmd.SetArgs([]string{"new", "myfullstack", "--preset", "fullstack"})
	err := rootCmd.ExecuteContext(context.Background())
	if err != nil {
		t.Fatalf("loy new failed: %v", err)
	}

	targetDir, _ := filesystem.CleanAndValidatePath(".", "myfullstack")

	expectedFiles := []string{
		targetDir + "/loy.yaml",
		targetDir + "/go.mod",
		targetDir + "/cmd/web/main.go",
		targetDir + "/internal/config/config.go",
		targetDir + "/internal/platform/logger/logger.go",
		targetDir + "/internal/platform/health/health.go",
		targetDir + "/internal/platform/shutdown/coordinator.go",
		targetDir + "/internal/app/app.go",
		targetDir + "/internal/app/wiring.go",
		targetDir + "/internal/transport/http/view/render.go",
		targetDir + "/internal/transport/http/view/assets.go",
		targetDir + "/internal/transport/http/handler/web.go",
		targetDir + "/internal/transport/http/static/static.go",
		targetDir + "/views/layouts/base.templ",
		targetDir + "/views/components/navbar.templ",
		targetDir + "/views/components/alert.templ",
		targetDir + "/views/pages/home.templ",
		targetDir + "/package.json",
		targetDir + "/vite.config.ts",
		targetDir + "/tailwind.config.js",
		targetDir + "/postcss.config.js",
		targetDir + "/assets/css/app.css",
		targetDir + "/assets/js/app.ts",
	}

	for _, file := range expectedFiles {
		exists, _ := memFS.Exists(file)
		if !exists {
			t.Errorf("expected file %s to be created by loy new --preset fullstack", file)
		}
	}

	// Verify cmd/api does NOT exist, since entrypoint is cmd/web
	apiExists, _ := memFS.Exists(targetDir + "/cmd/api/main.go")
	if apiExists {
		t.Errorf("expected cmd/api/main.go NOT to exist in fullstack preset")
	}

	wiringBytes, _ := memFS.ReadFile(targetDir + "/internal/app/wiring.go")
	if !strings.Contains(string(wiringBytes), "webHandler := handler.NewWebHandler()") {
		t.Errorf("expected webHandler in wiring.go:\n%s", string(wiringBytes))
	}
}

func TestNewProjectGeneratesWebRuntime(t *testing.T) {
	memFS := filesystem.NewMemFileSystem()
	rootCmd := cli.NewRootCmdWithFS(memFS, nil)

	rootCmd.SetArgs([]string{"new", "myweb", "--preset", "web"})
	err := rootCmd.ExecuteContext(context.Background())
	if err != nil {
		t.Fatalf("loy new failed: %v", err)
	}

	targetDir, _ := filesystem.CleanAndValidatePath(".", "myweb")

	// web preset has Templ but not Vite
	if exists, _ := memFS.Exists(targetDir + "/cmd/web/main.go"); !exists {
		t.Errorf("expected cmd/web/main.go to exist in web preset")
	}
	if exists, _ := memFS.Exists(targetDir + "/views/pages/home.templ"); !exists {
		t.Errorf("expected views/pages/home.templ to exist in web preset")
	}
	if exists, _ := memFS.Exists(targetDir + "/package.json"); exists {
		t.Errorf("expected package.json NOT to exist in web preset")
	}
}

func TestNewProject_ChiTransport(t *testing.T) {
	memFS := filesystem.NewMemFileSystem()
	rootCmd := cli.NewRootCmdWithFS(memFS, nil)

	rootCmd.SetArgs([]string{"new", "mychi", "--http", "chi"})
	err := rootCmd.ExecuteContext(context.Background())
	if err != nil {
		t.Fatalf("loy new --http chi failed: %v", err)
	}

	targetDir, _ := filesystem.CleanAndValidatePath(".", "mychi")
	appBytes, err := memFS.ReadFile(targetDir + "/internal/app/app.go")
	if err != nil {
		t.Fatalf("reading app.go: %v", err)
	}
	if !strings.Contains(string(appBytes), "github.com/go-chi/chi/v5") {
		t.Errorf("expected chi import in app.go")
	}

	serverBytes, err := memFS.ReadFile(targetDir + "/internal/transport/http/server.go")
	if err != nil {
		t.Fatalf("reading server.go: %v", err)
	}
	if !strings.Contains(string(serverBytes), "chi.NewRouter()") {
		t.Errorf("expected chi.NewRouter() in server.go")
	}
}

func TestNewProject_NetHTTPTransport(t *testing.T) {
	memFS := filesystem.NewMemFileSystem()
	rootCmd := cli.NewRootCmdWithFS(memFS, nil)

	rootCmd.SetArgs([]string{"new", "mynethttp", "--http", "nethttp"})
	err := rootCmd.ExecuteContext(context.Background())
	if err != nil {
		t.Fatalf("loy new --http nethttp failed: %v", err)
	}

	targetDir, _ := filesystem.CleanAndValidatePath(".", "mynethttp")
	appBytes, err := memFS.ReadFile(targetDir + "/internal/app/app.go")
	if err != nil {
		t.Fatalf("reading app.go: %v", err)
	}
	if !strings.Contains(string(appBytes), "http.ServeMux") {
		t.Errorf("expected http.ServeMux in app.go")
	}

	serverBytes, err := memFS.ReadFile(targetDir + "/internal/transport/http/server.go")
	if err != nil {
		t.Fatalf("reading server.go: %v", err)
	}
	if !strings.Contains(string(serverBytes), "http.NewServeMux()") {
		t.Errorf("expected http.NewServeMux() in server.go")
	}
}

func TestNewProject_GinTransport(t *testing.T) {
	memFS := filesystem.NewMemFileSystem()
	rootCmd := cli.NewRootCmdWithFS(memFS, nil)

	rootCmd.SetArgs([]string{"new", "mygin", "--http", "gin"})
	err := rootCmd.ExecuteContext(context.Background())
	if err != nil {
		t.Fatalf("loy new --http gin failed: %v", err)
	}

	targetDir, _ := filesystem.CleanAndValidatePath(".", "mygin")
	appBytes, err := memFS.ReadFile(targetDir + "/internal/app/app.go")
	if err != nil {
		t.Fatalf("reading app.go: %v", err)
	}
	if !strings.Contains(string(appBytes), "github.com/gin-gonic/gin") {
		t.Errorf("expected gin import in app.go")
	}

	serverBytes, err := memFS.ReadFile(targetDir + "/internal/transport/http/server.go")
	if err != nil {
		t.Fatalf("reading server.go: %v", err)
	}
	if !strings.Contains(string(serverBytes), "gin.New()") {
		t.Errorf("expected gin.New() in server.go")
	}
}

func TestNewProject_SqliteDatabase(t *testing.T) {
	memFS := filesystem.NewMemFileSystem()
	rootCmd := cli.NewRootCmdWithFS(memFS, nil)

	rootCmd.SetArgs([]string{"new", "mysqapp", "--db", "sqlite"})
	err := rootCmd.ExecuteContext(context.Background())
	if err != nil {
		t.Fatalf("loy new --db sqlite failed: %v", err)
	}

	targetDir, _ := filesystem.CleanAndValidatePath(".", "mysqapp")
	exists, _ := memFS.Exists(targetDir + "/internal/platform/database/sqlite.go")
	if !exists {
		t.Errorf("expected sqlite.go to be generated")
	}

	pgExists, _ := memFS.Exists(targetDir + "/internal/platform/database/postgres.go")
	if pgExists {
		t.Errorf("did not expect postgres.go when sqlite is selected")
	}
}

func TestNewProject_MySQLDatabase(t *testing.T) {
	memFS := filesystem.NewMemFileSystem()
	rootCmd := cli.NewRootCmdWithFS(memFS, nil)

	rootCmd.SetArgs([]string{"new", "mymyapp", "--db", "mysql"})
	err := rootCmd.ExecuteContext(context.Background())
	if err != nil {
		t.Fatalf("loy new --db mysql failed: %v", err)
	}

	targetDir, _ := filesystem.CleanAndValidatePath(".", "mymyapp")
	exists, _ := memFS.Exists(targetDir + "/internal/platform/database/mysql.go")
	if !exists {
		t.Errorf("expected mysql.go to be generated")
	}
}

func TestNewProject_InteractiveWizard(t *testing.T) {
	memFS := filesystem.NewMemFileSystem()
	rootCmd := cli.NewRootCmdWithFS(memFS, nil)

	// Simulate user typing:
	// Project name: "wizardapp"
	// Preset: (default: api)
	// HTTP: "chi"
	// DB: "sqlite"
	// Cache: (default)
	// Queue: (default)
	input := strings.NewReader("wizardapp\n\nchi\nsqlite\n\n\n")
	rootCmd.SetIn(input)
	rootCmd.SetArgs([]string{"new", "--interactive"})

	err := rootCmd.ExecuteContext(context.Background())
	if err != nil {
		t.Fatalf("loy new --interactive failed: %v", err)
	}

	targetDir, _ := filesystem.CleanAndValidatePath(".", "wizardapp")
	serverBytes, err := memFS.ReadFile(targetDir + "/internal/transport/http/server.go")
	if err != nil {
		t.Fatalf("reading server.go: %v", err)
	}
	if !strings.Contains(string(serverBytes), "chi.NewRouter()") {
		t.Errorf("expected chi router from wizard choice")
	}

	sqliteExists, _ := memFS.Exists(targetDir + "/internal/platform/database/sqlite.go")
	if !sqliteExists {
		t.Errorf("expected sqlite.go from wizard choice")
	}
}
