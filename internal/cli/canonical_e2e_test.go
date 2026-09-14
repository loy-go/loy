package cli_test

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/loy-go/loy/internal/cli"
)

func TestCanonicalEndToEndLoop(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping canonical end-to-end integration test in short mode")
	}

	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(origDir)
	})

	tmpDir := t.TempDir()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("chdir tmpDir: %v", err)
	}

	// 1. loy new demoapp --preset api
	root := cli.NewRootCmd()
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"new", "demoapp", "--preset", "api"})
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
	root.SetArgs([]string{"make", "crud", "products", "title:string:required", "price:int:required"})
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
		if _, err := os.Stat(filepath.Join(appDir, f)); os.IsNotExist(err) {
			t.Errorf("expected CRUD artifact %s to exist", f)
		}
	}

	// 3. loy make deploy all
	buf.Reset()
	root = cli.NewRootCmd()
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"make", "deploy", "all", "--force"})
	if err := root.ExecuteContext(context.Background()); err != nil {
		t.Fatalf("loy make deploy all failed: %v, output: %s", err, buf.String())
	}

	// Verify deployment artifacts
	expectedDeploy := []string{
		"Dockerfile",
		".dockerignore",
		"docker-compose.yml",
		"deploy/k8s/deployment.yaml",
		"deploy/k8s/service.yaml",
		"deploy/k8s/configmap.yaml",
		"deploy/k8s/ingress.yaml",
		"deploy/k8s/secret.yaml.example",
		"deploy/helm/demoapp/Chart.yaml",
		"deploy/helm/demoapp/values.yaml",
		".github/workflows/ci.yml",
	}
	for _, f := range expectedDeploy {
		if _, err := os.Stat(filepath.Join(appDir, f)); os.IsNotExist(err) {
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
	cmdBuild := exec.Command("go", "build", "./...")
	cmdBuild.Dir = appDir
	if out, err := cmdBuild.CombinedOutput(); err != nil {
		t.Fatalf("go build failed on generated app: %v\nOutput:\n%s", err, string(out))
	}
}
