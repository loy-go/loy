package workspace_test

import (
	"testing"

	"github.com/uloydev/loy/internal/diagnostics"
	"github.com/uloydev/loy/internal/filesystem"
	"github.com/uloydev/loy/internal/workspace"
)

func TestWorkspaceResolver_SingleModule(t *testing.T) {
	fs := filesystem.NewMemFileSystem()
	_ = fs.MkdirAll("/proj/cmd", 0755)
	_ = fs.WriteFile("/proj/go.mod", []byte("module github.com/acme/app\n\ngo 1.22\n\nrequire github.com/spf13/cobra v1.8.0\nrequire github.com/stretchr/testify v1.9.0 // indirect\n"), 0644)
	_ = fs.WriteFile("/proj/cmd/main.go", []byte("package main\nfunc main() {}\n"), 0644)

	res, err := workspace.NewResolver(fs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ws, diags := res.Resolve("/proj")
	if len(diags) > 0 {
		t.Fatalf("unexpected diagnostics: %+v", diags)
	}

	if len(ws.Apps) != 1 {
		t.Fatalf("expected 1 app, got %d", len(ws.Apps))
	}
	app := ws.Apps[0]
	if app.Name != "github.com/acme/app" || app.Type != workspace.TypeApp {
		t.Errorf("unexpected module details: %+v", app)
	}
	if len(app.DirectDependencies) != 1 || app.DirectDependencies[0] != "github.com/spf13/cobra" {
		t.Errorf("expected direct dependencies [github.com/spf13/cobra], got: %+v", app.DirectDependencies)
	}

	tgt, diag := ws.SelectTarget("")
	if diag != nil {
		t.Fatalf("unexpected error selecting default target: %+v", diag)
	}
	if tgt.Name != "github.com/acme/app" {
		t.Errorf("expected target github.com/acme/app, got %s", tgt.Name)
	}
}

func TestWorkspaceResolver_MultiModule(t *testing.T) {
	fs := filesystem.NewMemFileSystem()
	_ = fs.MkdirAll("/monorepo/apps/api", 0755)
	_ = fs.MkdirAll("/monorepo/apps/worker", 0755)
	_ = fs.MkdirAll("/monorepo/packages/auth", 0755)

	_ = fs.WriteFile("/monorepo/go.work", []byte("go 1.22\n\nuse (\n\t./apps/api\n\t./apps/worker\n\t./packages/auth\n)\n"), 0644)
	_ = fs.WriteFile("/monorepo/apps/api/go.mod", []byte("module github.com/acme/api\n\ngo 1.22\n"), 0644)
	_ = fs.WriteFile("/monorepo/apps/worker/go.mod", []byte("module github.com/acme/worker\n\ngo 1.22\n"), 0644)
	_ = fs.WriteFile("/monorepo/packages/auth/go.mod", []byte("module github.com/acme/auth\n\ngo 1.22\n"), 0644)

	res, err := workspace.NewResolver(fs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ws, diags := res.Resolve("/monorepo")
	if len(diags) > 0 {
		t.Fatalf("unexpected diagnostics: %+v", diags)
	}

	if len(ws.Apps) != 2 {
		t.Errorf("expected 2 apps, got %d", len(ws.Apps))
	}
	if len(ws.Packages) != 1 {
		t.Errorf("expected 1 package, got %d", len(ws.Packages))
	}

	// Ambiguous target without explicit selection
	_, diag := ws.SelectTarget("")
	if diag == nil || diag.Code != diagnostics.CodeWorkspaceTargetError {
		t.Errorf("expected error selecting default target with multiple apps, got %+v", diag)
	}

	// Select specific target
	tgt, diag := ws.SelectTarget("api")
	if diag != nil {
		t.Fatalf("failed to select 'api' target: %+v", diag)
	}
	if tgt.Name != "github.com/acme/api" {
		t.Errorf("expected api target, got %s", tgt.Name)
	}

	// Target not found
	_, diagNotFound := ws.SelectTarget("nonexistent")
	if diagNotFound == nil || diagNotFound.Code != diagnostics.CodeWorkspaceTargetError {
		t.Errorf("expected error selecting nonexistent target, got %+v", diagNotFound)
	}
}

func TestWorkspaceResolver_PathJailTraversal(t *testing.T) {
	fs := filesystem.NewMemFileSystem()
	_ = fs.MkdirAll("/monorepo", 0755)
	_ = fs.WriteFile("/monorepo/go.work", []byte("go 1.22\n\nuse (\n\t../outside\n)\n"), 0644)

	res, _ := workspace.NewResolver(fs)
	_, diags := res.Resolve("/monorepo")
	if len(diags) == 0 || diags[0].Code != diagnostics.CodeFSPathTraversal {
		t.Fatalf("expected path traversal diagnostic, got: %+v", diags)
	}
}

func TestWorkspaceResolver_Errors(t *testing.T) {
	_, err := workspace.NewResolver(nil)
	if err == nil {
		t.Fatal("expected error with nil filesystem")
	}

	fs := filesystem.NewMemFileSystem()
	res, _ := workspace.NewResolver(fs)

	// Neither go.work nor go.mod
	_, diags := res.Resolve("/empty")
	if len(diags) == 0 || diags[0].Code != diagnostics.CodeProjectRootNotFound {
		t.Fatalf("expected root not found diagnostic, got %+v", diags)
	}

	// Corrupt go.work
	_ = fs.WriteFile("/corrupt/go.work", []byte("invalid syntax"), 0644)
	_, diags = res.Resolve("/corrupt")
	if len(diags) == 0 || diags[0].Code != diagnostics.CodeWorkspaceConflict {
		t.Fatalf("expected conflict diagnostic for corrupt go.work, got %+v", diags)
	}
}
