package cli_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/loy-go/loy/internal/cli"
)

func executeMakeCommand(root *cobra.Command, args ...string) (string, error) {
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs(args)
	err := root.ExecuteContext(context.Background())
	return buf.String(), err
}

func TestMakeCommand(t *testing.T) {
	tempDir := t.TempDir()
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd error: %v", err)
	}
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("chdir error: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(origDir)
	})

	// Setup basic go.mod
	if err := os.WriteFile("go.mod", []byte("module github.com/example/testapp\n\ngo 1.22\n"), 0644); err != nil {
		t.Fatalf("writing go.mod error: %v", err)
	}

	root := cli.NewRootCmd()

	t.Run("make model", func(t *testing.T) {
		out, err := executeMakeCommand(root, "make", "model", "article", "title:string", "views:int")
		if err != nil {
			t.Fatalf("make model failed: %v, out: %s", err, out)
		}
		path := filepath.Join(tempDir, "internal/article/domain/article.go")
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Fatalf("expected file %s to exist", path)
		}
	})

	t.Run("make repo alias", func(t *testing.T) {
		out, err := executeMakeCommand(root, "make", "repo", "article")
		if err != nil {
			t.Fatalf("make repo failed: %v, out: %s", err, out)
		}
		path1 := filepath.Join(tempDir, "internal/article/domain/repository.go")
		path2 := filepath.Join(tempDir, "internal/article/repository/pg_adapter.go")
		if _, err := os.Stat(path1); os.IsNotExist(err) {
			t.Fatalf("expected file %s to exist", path1)
		}
		if _, err := os.Stat(path2); os.IsNotExist(err) {
			t.Fatalf("expected file %s to exist", path2)
		}
	})

	t.Run("make svc alias", func(t *testing.T) {
		out, err := executeMakeCommand(root, "make", "svc", "article")
		if err != nil {
			t.Fatalf("make svc failed: %v, out: %s", err, out)
		}
		path := filepath.Join(tempDir, "internal/article/service/service.go")
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Fatalf("expected file %s to exist", path)
		}
	})

	t.Run("make handler", func(t *testing.T) {
		out, err := executeMakeCommand(root, "make", "handler", "article")
		if err != nil {
			t.Fatalf("make handler failed: %v, out: %s", err, out)
		}
		path := filepath.Join(tempDir, "internal/article/transport/http/handler.go")
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Fatalf("expected file %s to exist", path)
		}
	})

	t.Run("make feature with auto wiring", func(t *testing.T) {
		out, err := executeMakeCommand(root, "make", "feature", "comment", "body:string:required")
		if err != nil {
			t.Fatalf("make feature failed: %v, out: %s", err, out)
		}
		p1 := filepath.Join(tempDir, "internal/comment/domain/comment.go")
		p2 := filepath.Join(tempDir, "internal/comment/service/service.go")
		p3 := filepath.Join(tempDir, "internal/comment/transport/http/handler.go")
		pw := filepath.Join(tempDir, "internal/app/wiring.go")

		for _, p := range []string{p1, p2, p3, pw} {
			if _, err := os.Stat(p); os.IsNotExist(err) {
				t.Fatalf("expected file %s to exist", p)
			}
		}

		wiringBytes, err := os.ReadFile(pw)
		if err != nil {
			t.Fatalf("reading wiring.go error: %v", err)
		}
		wiringContent := string(wiringBytes)
		if !strings.Contains(wiringContent, "commentRepo") {
			t.Fatalf("missing commentRepo in wiring: %s", wiringContent)
		}
		if !strings.Contains(wiringContent, "commentSvc") {
			t.Fatalf("missing commentSvc in wiring: %s", wiringContent)
		}
		if !strings.Contains(wiringContent, "commentHandler") {
			t.Fatalf("missing commentHandler in wiring: %s", wiringContent)
		}
	})

	t.Run("make tenant", func(t *testing.T) {
		out, err := executeMakeCommand(root, "make", "tenant", "tenant")
		if err != nil {
			t.Fatalf("make tenant failed: %v, out: %s", err, out)
		}
		tenantPath := filepath.Join(tempDir, "internal/platform/tenant/context.go")
		if _, err := os.Stat(tenantPath); os.IsNotExist(err) {
			t.Fatalf("expected file %s to exist, output: %s", tenantPath, out)
		}
	})

	t.Run("make dry run", func(t *testing.T) {
		out, err := executeMakeCommand(root, "make", "job", "email_sender", "--dry-run")
		if err != nil {
			t.Fatalf("make dry-run failed: %v, out: %s", err, out)
		}
		if !strings.Contains(out, "Plan operations (dry-run)") {
			t.Fatalf("expected dry-run header, got %s", out)
		}
		path := filepath.Join(tempDir, "internal/emailsender/job/email_sender_job.go")
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Fatalf("expected file %s not to exist during dry-run", path)
		}
	})

	t.Run("make view full page", func(t *testing.T) {
		r := cli.NewRootCmd()
		out, err := executeMakeCommand(r, "make", "view", "dashboard")
		if err != nil {
			t.Fatalf("make view failed: %v, out: %s", err, out)
		}
		path := filepath.Join(tempDir, "views/pages/dashboard.templ")
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Fatalf("expected file %s to exist, output: %s", path, out)
		}
		content, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("reading generated view: %v", err)
		}
		if !strings.Contains(string(content), "templ Dashboard()") {
			t.Fatalf("expected templ Dashboard(), got:\n%s", string(content))
		}
	})

	t.Run("make view partial component", func(t *testing.T) {
		r := cli.NewRootCmd()
		out, err := executeMakeCommand(r, "make", "view", "user_row", "--partial")
		if err != nil {
			t.Fatalf("make view --partial failed: %v, out: %s", err, out)
		}
		path := filepath.Join(tempDir, "views/components/user_row.templ")
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Fatalf("expected file %s to exist", path)
		}
		content, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("reading generated partial view: %v", err)
		}
		if !strings.Contains(string(content), "templ UserRow()") {
			t.Fatalf("expected templ UserRow(), got:\n%s", string(content))
		}
		if !strings.Contains(string(content), "hx-target") {
			t.Fatalf("expected hx-target, got:\n%s", string(content))
		}
	})

	t.Run("make docker", func(t *testing.T) {
		r := cli.NewRootCmd()
		out, err := executeMakeCommand(r, "make", "docker")
		if err != nil {
			t.Fatalf("make docker failed: %v, out: %s", err, out)
		}
		if _, err := os.Stat(filepath.Join(tempDir, "Dockerfile")); os.IsNotExist(err) {
			t.Fatalf("expected Dockerfile to exist")
		}
		if _, err := os.Stat(filepath.Join(tempDir, "docker-compose.yml")); os.IsNotExist(err) {
			t.Fatalf("expected docker-compose.yml to exist")
		}
	})

	t.Run("make k8s", func(t *testing.T) {
		r := cli.NewRootCmd()
		out, err := executeMakeCommand(r, "make", "k8s")
		if err != nil {
			t.Fatalf("make k8s failed: %v, out: %s", err, out)
		}
		if _, err := os.Stat(filepath.Join(tempDir, "deploy/k8s/deployment.yaml")); os.IsNotExist(err) {
			t.Fatalf("expected deploy/k8s/deployment.yaml to exist")
		}
		if _, err := os.Stat(filepath.Join(tempDir, "deploy/k8s/service.yaml")); os.IsNotExist(err) {
			t.Fatalf("expected deploy/k8s/service.yaml to exist")
		}
	})

	t.Run("make helm", func(t *testing.T) {
		r := cli.NewRootCmd()
		out, err := executeMakeCommand(r, "make", "helm", "testapp")
		if err != nil {
			t.Fatalf("make helm failed: %v, out: %s", err, out)
		}
		if _, err := os.Stat(filepath.Join(tempDir, "deploy/helm/testapp/Chart.yaml")); os.IsNotExist(err) {
			t.Fatalf("expected deploy/helm/testapp/Chart.yaml to exist")
		}
	})

	t.Run("make ci", func(t *testing.T) {
		r := cli.NewRootCmd()
		out, err := executeMakeCommand(r, "make", "ci")
		if err != nil {
			t.Fatalf("make ci failed: %v, out: %s", err, out)
		}
		if _, err := os.Stat(filepath.Join(tempDir, ".github/workflows/ci.yml")); os.IsNotExist(err) {
			t.Fatalf("expected .github/workflows/ci.yml to exist")
		}
	})

	t.Run("make deploy all", func(t *testing.T) {
		r := cli.NewRootCmd()
		out, err := executeMakeCommand(r, "make", "deploy", "--force")
		if err != nil {
			t.Fatalf("make deploy failed: %v, out: %s", err, out)
		}
		expected := []string{
			"Dockerfile",
			"docker-compose.yml",
			"deploy/k8s/deployment.yaml",
			".github/workflows/ci.yml",
		}
		for _, f := range expected {
			if _, err := os.Stat(filepath.Join(tempDir, f)); os.IsNotExist(err) {
				t.Errorf("expected %s to exist after make deploy", f)
			}
		}
	})
}
