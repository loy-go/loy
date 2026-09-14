package cli_test

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/uloydev/loy/internal/cli"
	"github.com/uloydev/loy/internal/filesystem"
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
	if err == nil {
		t.Cleanup(func() {
			_ = os.Chdir(origDir)
		})
	}
	tmpDir := t.TempDir()
	memFS := filesystem.NewOSFileSystem()
	_ = memFS.MkdirAll(tmpDir, 0755)
	_ = memFS.WriteFile(tmpDir+"/go.mod", []byte("module github.com/test/workspace\n\ngo 1.22\n"), 0644)
	_ = memFS.WriteFile(tmpDir+"/loy.yaml", []byte("version: 1\nproject:\n  name: workspace\n"), 0644)

	rootCmd := cli.NewRootCmdWithFS(memFS, nil)
	rootCmd.SetArgs([]string{"make", "runtime", "--http", "nethttp", "-C", tmpDir})
	err = rootCmd.ExecuteContext(context.Background())
	if err != nil {
		t.Fatalf("loy make runtime failed: %v", err)
	}

	exists, _ := memFS.Exists(tmpDir + "/internal/app/app.go")
	if !exists {
		t.Fatalf("expected internal/app/app.go to be generated")
	}

	content, _ := memFS.ReadFile(tmpDir + "/internal/app/app.go")
	if !strings.Contains(string(content), "http.ServeMux") {
		t.Errorf("expected net/http server in app.go")
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
