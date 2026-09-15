package cli_test

import (
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/loy-go/loy/internal/cli"
	"github.com/loy-go/loy/internal/filesystem"
	"github.com/loy-go/loy/internal/process"
)

type pluginMockRunner struct {
	stdout   []byte
	stderr   []byte
	exitCode int
}

func (m *pluginMockRunner) Run(ctx context.Context, dir string, name string, args ...string) (*process.Result, error) {
	if name == "go" {
		return &process.Result{ExitCode: 1}, nil
	}
	return &process.Result{
		Stdout:   m.stdout,
		Stderr:   m.stderr,
		ExitCode: m.exitCode,
	}, nil
}

func (m *pluginMockRunner) RunWithInput(ctx context.Context, dir string, in io.Reader, name string, args ...string) (*process.Result, error) {
	return &process.Result{
		Stdout:   m.stdout,
		Stderr:   m.stderr,
		ExitCode: m.exitCode,
	}, nil
}

func TestPluginCLICommands(t *testing.T) {
	tempDir := t.TempDir()
	origDir, _ := os.Getwd()
	_ = os.Chdir(tempDir)
	t.Cleanup(func() { _ = os.Chdir(origDir) })

	memFS := filesystem.NewMemFileSystem()
	runner := &pluginMockRunner{}

	projDir := filepath.Join(tempDir, "proj")
	_ = os.MkdirAll(projDir, 0755)
	_ = memFS.MkdirAll(projDir, 0755)
	_ = memFS.WriteFile(filepath.Join(projDir, "loy.yaml"), []byte("version: 1\nproject:\n  name: proj\n"), 0644)

	// 1. loy plugin list when empty
	buf := new(bytes.Buffer)
	rootCmd := cli.NewRootCmdWithFS(memFS, runner)
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"plugin", "list", projDir})

	err := rootCmd.ExecuteContext(context.Background())
	if err != nil {
		t.Fatalf("plugin list failed: %v", err)
	}
	if !strings.Contains(buf.String(), "No plugins installed") {
		t.Errorf("expected 'No plugins installed', got: %s", buf.String())
	}

	// 2. Setup mock plugin source and install
	srcDir := filepath.Join(tempDir, "src", "stripe")
	_ = memFS.MkdirAll(srcDir, 0755)
	_ = memFS.WriteFile(filepath.Join(srcDir, "loy-plugin.yaml"), []byte(`name: stripe
version: 1.0.0
description: Stripe webhook generator
commands:
  - name: webhook
    description: Generate webhook handler
    executable: bin/stripe
`), 0644)

	buf.Reset()
	rootCmd = cli.NewRootCmdWithFS(memFS, runner)
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"plugin", "install", srcDir, projDir})

	err = rootCmd.ExecuteContext(context.Background())
	if err != nil {
		t.Fatalf("plugin install failed: %v", err)
	}
	if !strings.Contains(buf.String(), "Successfully installed plugin \"stripe\"") {
		t.Errorf("expected successful install message, got: %s", buf.String())
	}

	// 3. loy plugin list with installed plugin
	buf.Reset()
	rootCmd = cli.NewRootCmdWithFS(memFS, runner)
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"plugin", "list", projDir})

	err = rootCmd.ExecuteContext(context.Background())
	if err != nil {
		t.Fatalf("plugin list failed: %v", err)
	}
	if !strings.Contains(buf.String(), "stripe (v1.0.0)") || !strings.Contains(buf.String(), "webhook") {
		t.Errorf("expected stripe plugin in list, got: %s", buf.String())
	}

	// 4. loy plugin run stripe webhook
	runner.stdout = []byte(`{
  "status": "ok",
  "artifacts": [
    {
      "path": "internal/payment/webhook.go",
      "content": "package payment\n\n// stripe webhook\n",
      "ownership": "developer"
    }
  ]
}`)

	buf.Reset()
	rootCmd = cli.NewRootCmdWithFS(memFS, runner)
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"plugin", "run", "stripe", "webhook", "-C", projDir})

	err = rootCmd.ExecuteContext(context.Background())
	if err != nil {
		t.Fatalf("plugin run failed: %v", err)
	}
	if !strings.Contains(buf.String(), "Successfully generated 1 artifact(s)") {
		t.Errorf("expected generated message, got: %s", buf.String())
	}

	// Verify artifact on disk
	exists, _ := memFS.Exists(filepath.Join(projDir, "internal/payment/webhook.go"))
	if !exists {
		t.Errorf("expected generated file in project")
	}
}
