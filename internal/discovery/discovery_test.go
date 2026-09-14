package discovery_test

import (
	"context"
	"testing"

	"github.com/loy-go/loy/internal/diagnostics"
	"github.com/loy-go/loy/internal/discovery"
	"github.com/loy-go/loy/internal/filesystem"
)

func TestDiscoverer_FindsLoyManifest(t *testing.T) {
	fs := filesystem.NewMemFileSystem()
	_ = fs.MkdirAll("/repo/apps/api/src", 0755)
	_ = fs.WriteFile("/repo/loy.yaml", []byte("version: 1\nproject:\n  name: repo\n"), 0644)
	_ = fs.WriteFile("/repo/go.mod", []byte("module repo\n"), 0644)

	disc, err := discovery.NewDiscoverer(fs, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ctx := context.Background()
	res, diag := disc.Discover(ctx, "/repo/apps/api/src")
	if diag != nil {
		t.Fatalf("unexpected diagnostic: %+v", diag)
	}

	if res.RootDir != "/repo" {
		t.Errorf("expected root /repo, got %s", res.RootDir)
	}
	if !res.HasManifest || !res.HasGoMod {
		t.Errorf("expected manifest and go.mod detected, got: %+v", res)
	}
}

func TestDiscoverer_FindsGoWork(t *testing.T) {
	fs := filesystem.NewMemFileSystem()
	_ = fs.MkdirAll("/workspace/apps/web", 0755)
	_ = fs.WriteFile("/workspace/go.work", []byte("go 1.22\nuse (\n\t./apps/web\n)\n"), 0644)
	_ = fs.WriteFile("/workspace/apps/web/go.mod", []byte("module web\n"), 0644)

	disc, err := discovery.NewDiscoverer(fs, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ctx := context.Background()
	res, diag := disc.Discover(ctx, "/workspace/apps/web")
	if diag != nil {
		t.Fatalf("unexpected diagnostic: %+v", diag)
	}

	if res.RootDir != "/workspace/apps/web" {
		t.Errorf("expected project root /workspace/apps/web, got %s", res.RootDir)
	}
	if !res.IsWorkspace || !res.HasGoWork {
		t.Errorf("expected workspace detected, got: %+v", res)
	}
	if res.GoWorkPath != "/workspace/go.work" {
		t.Errorf("expected go.work path /workspace/go.work, got %s", res.GoWorkPath)
	}
}

func TestDiscoverer_InferProjectName(t *testing.T) {
	fs := filesystem.NewMemFileSystem()
	_ = fs.MkdirAll("/myproject", 0755)
	_ = fs.WriteFile("/myproject/go.mod", []byte("module github.com/example/super-api\n\ngo 1.22\n"), 0644)

	disc, _ := discovery.NewDiscoverer(fs, nil)
	name, err := disc.InferProjectName("/myproject")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if name != "super-api" {
		t.Errorf("expected super-api, got %s", name)
	}
}

func TestDiscoverer_NotFound(t *testing.T) {
	fs := filesystem.NewMemFileSystem()
	_ = fs.MkdirAll("/empty/dir", 0755)

	disc, err := discovery.NewDiscoverer(fs, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ctx := context.Background()
	_, diag := disc.Discover(ctx, "/empty/dir")
	if diag == nil {
		t.Fatal("expected diagnostic for missing root")
	}
	if diag.Code != diagnostics.CodeProjectRootNotFound {
		t.Errorf("expected %s, got %s", diagnostics.CodeProjectRootNotFound, diag.Code)
	}
}
