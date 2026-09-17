package cli_test

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/loy-go/loy/internal/cli"
)

func setupE2EWorkspace(t *testing.T) (string, string) {
	t.Helper()
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}

	tmpDir := t.TempDir()
	t.Cleanup(func() {
		_ = os.Chdir(origDir)
		for i := 0; i < 15; i++ {
			if err := os.RemoveAll(tmpDir); err == nil || os.IsNotExist(err) {
				break
			}
			time.Sleep(100 * time.Millisecond)
		}
	})

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("chdir tmpDir: %v", err)
	}
	return tmpDir, origDir
}

func TestCanonicalEndToEndLoop(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping canonical end-to-end integration test in short mode")
	}

	tmpDir, _ := setupE2EWorkspace(t)

	// 1. loy new demoapp --preset api
	root := cli.NewRootCmd()
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"new", "demoapp", "--preset", "api", "--no-tidy"})
	if err := root.ExecuteContext(context.Background()); err != nil {
		t.Fatalf("loy new failed: %v, output: %s", err, buf.String())
	}

	appDir := filepath.Join(tmpDir, "demoapp")
	if err := os.Chdir(appDir); err != nil {
		t.Fatalf("chdir appDir: %v", err)
	}

	// 2. loy make crud products
	buf.Reset()
	root = cli.NewRootCmd()
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"make", "crud", "products", "title:string:required", "price:int:required", "--no-tidy"})
	if err := root.ExecuteContext(context.Background()); err != nil {
		t.Fatalf("loy make crud failed: %v, output: %s", err, buf.String())
	}

	// Verify CRUD artifacts
	expectedCRUD := []string{
		"internal/products/domain/products.go",
		"internal/products/domain/repository.go",
		"internal/products/repository/pg_adapter.go",
		"internal/products/service/service.go",
		"internal/products/transport/http/handler.go",
		"internal/app/wiring.go",
	}
	for _, f := range expectedCRUD {
		if _, err := os.Stat(filepath.Join(appDir, filepath.FromSlash(f))); os.IsNotExist(err) {
			t.Errorf("expected CRUD artifact %s to exist", f)
		}
	}

	// 3. loy make deploy all
	buf.Reset()
	root = cli.NewRootCmd()
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"make", "deploy", "all", "--force", "--no-tidy"})
	if err := root.ExecuteContext(context.Background()); err != nil {
		t.Fatalf("loy make deploy all failed: %v, output: %s", err, buf.String())
	}

	expectedDeploy := []string{
		"Dockerfile",
		"docker-compose.yml",
		"deploy/k8s/deployment.yaml",
		".github/workflows/ci.yml",
	}
	for _, f := range expectedDeploy {
		if _, err := os.Stat(filepath.Join(appDir, filepath.FromSlash(f))); os.IsNotExist(err) {
			t.Errorf("expected deployment artifact %s to exist", f)
		}
	}

	// 4. loy check (architecture verification)
	buf.Reset()
	root = cli.NewRootCmd()
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"check"})
	if err := root.ExecuteContext(context.Background()); err != nil {
		t.Fatalf("loy check failed on generated app: %v, output: %s", err, buf.String())
	}

	// 5. go mod tidy & go test ./...
	cmdTidy := exec.Command("go", "mod", "tidy")
	cmdTidy.Dir = appDir
	if out, err := cmdTidy.CombinedOutput(); err != nil {
		t.Fatalf("go mod tidy failed on generated app: %v\nOutput:\n%s", err, string(out))
	}

	cmdTest := exec.Command("go", "test", "./...")
	cmdTest.Dir = appDir
	if out, err := cmdTest.CombinedOutput(); err != nil {
		t.Fatalf("go test failed on generated app: %v\nOutput:\n%s", err, string(out))
	}

	// 6. go build ./...
	cmdBuild := exec.Command("go", "build", "-o", os.DevNull, "./...")
	cmdBuild.Dir = appDir
	if out, err := cmdBuild.CombinedOutput(); err != nil {
		t.Fatalf("go build failed on generated app: %v\nOutput:\n%s", err, string(out))
	}
}

func TestCanonicalEndToEndLoop_Chi(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping canonical end-to-end integration test in short mode")
	}

	tmpDir, _ := setupE2EWorkspace(t)

	// 1. loy new chiapp --preset api --http chi --db sqlite
	root := cli.NewRootCmd()
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"new", "chiapp", "--preset", "api", "--http", "chi", "--db", "sqlite", "--no-tidy"})
	if err := root.ExecuteContext(context.Background()); err != nil {
		t.Fatalf("loy new failed: %v, output: %s", err, buf.String())
	}

	appDir := filepath.Join(tmpDir, "chiapp")
	if err := os.Chdir(appDir); err != nil {
		t.Fatalf("chdir appDir: %v", err)
	}

	// 2. loy make crud tasks
	buf.Reset()
	root = cli.NewRootCmd()
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"make", "crud", "tasks", "name:string:required", "--no-tidy"})
	if err := root.ExecuteContext(context.Background()); err != nil {
		t.Fatalf("loy make crud failed: %v, output: %s", err, buf.String())
	}

	// 3. loy check
	buf.Reset()
	root = cli.NewRootCmd()
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"check"})
	if err := root.ExecuteContext(context.Background()); err != nil {
		t.Fatalf("loy check failed on generated app: %v, output: %s", err, buf.String())
	}

	// 4. go mod tidy & go test ./...
	cmdTidy := exec.Command("go", "mod", "tidy")
	cmdTidy.Dir = appDir
	if out, err := cmdTidy.CombinedOutput(); err != nil {
		t.Fatalf("go mod tidy failed on generated app: %v\nOutput:\n%s", err, string(out))
	}

	cmdTest := exec.Command("go", "test", "./...")
	cmdTest.Dir = appDir
	if out, err := cmdTest.CombinedOutput(); err != nil {
		t.Fatalf("go test failed on generated app: %v\nOutput:\n%s", err, string(out))
	}

	// 5. go build ./...
	cmdBuild := exec.Command("go", "build", "-o", os.DevNull, "./...")
	cmdBuild.Dir = appDir
	if out, err := cmdBuild.CombinedOutput(); err != nil {
		t.Fatalf("go build failed on generated app: %v\nOutput:\n%s", err, string(out))
	}
}

func TestCanonicalEndToEndLoop_NetHTTP(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping canonical end-to-end integration test in short mode")
	}

	tmpDir, _ := setupE2EWorkspace(t)

	// 1. loy new netapp --preset api --http nethttp --db sqlite
	root := cli.NewRootCmd()
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"new", "netapp", "--preset", "api", "--http", "nethttp", "--db", "sqlite", "--no-tidy"})
	if err := root.ExecuteContext(context.Background()); err != nil {
		t.Fatalf("loy new failed: %v, output: %s", err, buf.String())
	}

	appDir := filepath.Join(tmpDir, "netapp")
	if err := os.Chdir(appDir); err != nil {
		t.Fatalf("chdir appDir: %v", err)
	}

	// 2. loy make crud notes
	buf.Reset()
	root = cli.NewRootCmd()
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"make", "crud", "notes", "title:string:required", "--no-tidy"})
	if err := root.ExecuteContext(context.Background()); err != nil {
		t.Fatalf("loy make crud failed: %v, output: %s", err, buf.String())
	}

	// 3. loy check
	buf.Reset()
	root = cli.NewRootCmd()
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"check"})
	if err := root.ExecuteContext(context.Background()); err != nil {
		t.Fatalf("loy check failed on generated app: %v, output: %s", err, buf.String())
	}

	// 4. go mod tidy & go test ./...
	cmdTidy := exec.Command("go", "mod", "tidy")
	cmdTidy.Dir = appDir
	if out, err := cmdTidy.CombinedOutput(); err != nil {
		t.Fatalf("go mod tidy failed on generated app: %v\nOutput:\n%s", err, string(out))
	}

	cmdTest := exec.Command("go", "test", "./...")
	cmdTest.Dir = appDir
	if out, err := cmdTest.CombinedOutput(); err != nil {
		t.Fatalf("go test failed on generated app: %v\nOutput:\n%s", err, string(out))
	}

	// 5. go build ./...
	cmdBuild := exec.Command("go", "build", "-o", os.DevNull, "./...")
	cmdBuild.Dir = appDir
	if out, err := cmdBuild.CombinedOutput(); err != nil {
		t.Fatalf("go build failed on generated app: %v\nOutput:\n%s", err, string(out))
	}
}

func TestCanonicalEndToEndLoop_Gin(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping canonical end-to-end integration test in short mode")
	}

	tmpDir, _ := setupE2EWorkspace(t)

	// 1. loy new ginapp --preset api --http gin --db sqlite
	root := cli.NewRootCmd()
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"new", "ginapp", "--preset", "api", "--http", "gin", "--db", "sqlite", "--no-tidy"})
	if err := root.ExecuteContext(context.Background()); err != nil {
		t.Fatalf("loy new failed: %v, output: %s", err, buf.String())
	}

	appDir := filepath.Join(tmpDir, "ginapp")
	if err := os.Chdir(appDir); err != nil {
		t.Fatalf("chdir appDir: %v", err)
	}

	// 2. loy make crud articles
	buf.Reset()
	root = cli.NewRootCmd()
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"make", "crud", "articles", "title:string:required", "--no-tidy"})
	if err := root.ExecuteContext(context.Background()); err != nil {
		t.Fatalf("loy make crud failed: %v, output: %s", err, buf.String())
	}

	// 3. loy check
	buf.Reset()
	root = cli.NewRootCmd()
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"check"})
	if err := root.ExecuteContext(context.Background()); err != nil {
		t.Fatalf("loy check failed on generated app: %v, output: %s", err, buf.String())
	}

	// 4. go mod tidy & go test ./...
	cmdTidy := exec.Command("go", "mod", "tidy")
	cmdTidy.Dir = appDir
	if out, err := cmdTidy.CombinedOutput(); err != nil {
		t.Fatalf("go mod tidy failed on generated app: %v\nOutput:\n%s", err, string(out))
	}

	cmdTest := exec.Command("go", "test", "./...")
	cmdTest.Dir = appDir
	if out, err := cmdTest.CombinedOutput(); err != nil {
		t.Fatalf("go test failed on generated app: %v\nOutput:\n%s", err, string(out))
	}

	// 5. go build ./...
	cmdBuild := exec.Command("go", "build", "-o", os.DevNull, "./...")
	cmdBuild.Dir = appDir
	if out, err := cmdBuild.CombinedOutput(); err != nil {
		t.Fatalf("go build failed on generated app: %v\nOutput:\n%s", err, string(out))
	}
}
