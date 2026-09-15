package plugin_test

import (
	"context"
	"io"
	"strings"
	"testing"

	"github.com/loy-go/loy/internal/filesystem"
	"github.com/loy-go/loy/internal/plugin"
	"github.com/loy-go/loy/internal/process"
)

type mockRunner struct {
	stdout   []byte
	stderr   []byte
	exitCode int
}

func (m *mockRunner) Run(ctx context.Context, dir string, name string, args ...string) (*process.Result, error) {
	return &process.Result{
		Stdout:   m.stdout,
		Stderr:   m.stderr,
		ExitCode: m.exitCode,
	}, nil
}

func (m *mockRunner) RunWithInput(ctx context.Context, dir string, in io.Reader, name string, args ...string) (*process.Result, error) {
	return &process.Result{
		Stdout:   m.stdout,
		Stderr:   m.stderr,
		ExitCode: m.exitCode,
	}, nil
}

func TestParsePluginManifest(t *testing.T) {
	t.Run("valid manifest", func(t *testing.T) {
		data := []byte(`name: stripe
version: 1.0.0
description: Stripe webhook generator
commands:
  - name: webhook
    description: Generate webhook handler
    executable: bin/stripe
`)
		pm, err := plugin.ParsePluginManifest(data)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if pm.Name != "stripe" || len(pm.Commands) != 1 {
			t.Errorf("unexpected parsed manifest: %+v", pm)
		}
	})

	t.Run("missing name", func(t *testing.T) {
		data := []byte(`version: 1.0.0
commands:
  - name: webhook
    executable: bin/stripe
`)
		_, err := plugin.ParsePluginManifest(data)
		if err == nil {
			t.Errorf("expected error for missing name")
		}
	})

	t.Run("path traversal in executable rejected", func(t *testing.T) {
		data := []byte(`name: evil
version: 1.0.0
commands:
  - name: attack
    executable: ../../../bin/sh
`)
		_, err := plugin.ParsePluginManifest(data)
		if err == nil {
			t.Errorf("expected error for traversal in executable")
		}
	})
}

func TestPluginManager_Execute_Sandboxed(t *testing.T) {
	memFS := filesystem.NewMemFileSystem()
	ctx := context.Background()

	_ = memFS.MkdirAll("/app/.loy/plugins/stripe", 0755)
	_ = memFS.WriteFile("/app/.loy/plugins/stripe/loy-plugin.yaml", []byte(`name: stripe
version: 1.0.0
commands:
  - name: webhook
    executable: bin/stripe
`), 0644)

	t.Run("happy path execution", func(t *testing.T) {
		runner := &mockRunner{
			stdout: []byte(`{
  "status": "ok",
  "artifacts": [
    {
      "path": "internal/payment/webhook.go",
      "content": "package payment\n",
      "ownership": "developer"
    }
  ]
}`),
		}

		mgr := plugin.NewManager(memFS, runner, "/app")
		arts, err := mgr.Execute(ctx, "stripe", "webhook", nil, "github.com/example/app", nil, false, false)
		if err != nil {
			t.Fatalf("plugin execute failed: %v", err)
		}
		if len(arts) != 1 {
			t.Fatalf("expected 1 artifact, got %d", len(arts))
		}

		exists, _ := memFS.Exists("/app/internal/payment/webhook.go")
		if !exists {
			t.Errorf("expected artifact to be written to /app/internal/payment/webhook.go")
		}
	})

	t.Run("path traversal attempt blocked by sandbox", func(t *testing.T) {
		runner := &mockRunner{
			stdout: []byte(`{
  "status": "ok",
  "artifacts": [
    {
      "path": "../../etc/passwd",
      "content": "root:x:0:0...",
      "ownership": "developer"
    }
  ]
}`),
		}

		mgr := plugin.NewManager(memFS, runner, "/app")
		_, err := mgr.Execute(ctx, "stripe", "webhook", nil, "github.com/example/app", nil, false, false)
		if err == nil {
			t.Fatalf("expected security error when plugin emits path traversal")
		}
		if !strings.Contains(err.Error(), "security violation") {
			t.Errorf("expected security violation error, got: %v", err)
		}
	})

	t.Run(".git write attempt blocked by sandbox", func(t *testing.T) {
		runner := &mockRunner{
			stdout: []byte(`{
  "status": "ok",
  "artifacts": [
    {
      "path": ".git/hooks/pre-commit",
      "content": "#!/bin/sh\nexit 0",
      "ownership": "developer"
    }
  ]
}`),
		}

		mgr := plugin.NewManager(memFS, runner, "/app")
		_, err := mgr.Execute(ctx, "stripe", "webhook", nil, "github.com/example/app", nil, false, false)
		if err == nil {
			t.Fatalf("expected security error when plugin targets .git")
		}
		if !strings.Contains(err.Error(), "security violation") || !strings.Contains(err.Error(), ".git") {
			t.Errorf("expected .git security violation error, got: %v", err)
		}
	})
}

func TestPluginManager_ListAndInstall(t *testing.T) {
	memFS := filesystem.NewMemFileSystem()
	ctx := context.Background()

	// Create source plugin
	_ = memFS.MkdirAll("/local/myplugin", 0755)
	_ = memFS.WriteFile("/local/myplugin/loy-plugin.yaml", []byte(`name: myplugin
version: 2.0.0
commands:
  - name: gen
    executable: bin/gen
`), 0644)

	runner := &mockRunner{}
	mgr := plugin.NewManager(memFS, runner, "/app")

	// 1. Install
	pm, err := mgr.Install(ctx, "/local/myplugin")
	if err != nil {
		t.Fatalf("install failed: %v", err)
	}
	if pm.Name != "myplugin" {
		t.Errorf("expected myplugin, got %s", pm.Name)
	}

	// 2. List
	plugins, err := mgr.List(ctx)
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}
	if len(plugins) != 1 || plugins[0].Name != "myplugin" {
		t.Errorf("expected 1 plugin named myplugin, got: %v", plugins)
	}
}
