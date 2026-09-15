package cli_test

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/loy-go/loy/internal/cli"
	"github.com/loy-go/loy/internal/filesystem"
	"github.com/loy-go/loy/internal/process"
)

func TestUpgradeCommand_NoManifest(t *testing.T) {
	memFS := filesystem.NewMemFileSystem()
	execRunner := process.NewExecRunner()

	rootCmd := cli.NewRootCmdWithFS(memFS, execRunner)
	rootCmd.SetArgs([]string{"upgrade", "/empty"})

	err := rootCmd.ExecuteContext(context.Background())
	if err == nil {
		t.Fatalf("expected error when running upgrade on directory without loy.yaml")
	}
}

func TestUpgradeCommand_UpToDate(t *testing.T) {
	memFS := filesystem.NewMemFileSystem()
	execRunner := process.NewExecRunner()

	_ = memFS.MkdirAll("/proj/internal/app", 0755)
	_ = memFS.WriteFile("/proj/loy.yaml", []byte(`version: 1
project:
  name: proj
defaults:
  http: fiber
  database: postgres
`), 0644)
	_ = memFS.WriteFile("/proj/internal/app/wiring.go", []byte(`package app
import (
	// loy:region:imports
	// loy:endregion
)
func (a *App) wireDependencies() error {
	// loy:region:repositories
	// loy:endregion
	// loy:region:services
	// loy:endregion
	// loy:region:handlers
	// loy:endregion
	// loy:region:routes
	// loy:endregion
	return nil
}
`), 0644)

	buf := new(bytes.Buffer)
	rootCmd := cli.NewRootCmdWithFS(memFS, execRunner)
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"upgrade", "/proj"})

	err := rootCmd.ExecuteContext(context.Background())
	if err != nil {
		t.Fatalf("upgrade failed: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "up to date") {
		t.Errorf("expected 'up to date' message, got: %s", out)
	}
}

func TestUpgradeCommand_PlanAndApply(t *testing.T) {
	memFS := filesystem.NewMemFileSystem()
	execRunner := process.NewExecRunner()

	// Setup legacy or drifted project:
	// - missing version in loy.yaml
	// - missing defaults in loy.yaml
	// - missing // loy:region:routes in wiring.go
	_ = memFS.MkdirAll("/legacy/internal/app", 0755)
	_ = memFS.WriteFile("/legacy/loy.yaml", []byte(`project:
  name: legacy
`), 0644)
	_ = memFS.WriteFile("/legacy/internal/app/wiring.go", []byte(`package app
import (
	// loy:region:imports
	// loy:endregion
)
func (a *App) wireDependencies() error {
	// loy:region:repositories
	// loy:endregion
	// loy:region:services
	// loy:endregion
	// loy:region:handlers
	// loy:endregion
	return nil
}
`), 0644)

	// 1. Test Dry Run
	dryBuf := new(bytes.Buffer)
	rootCmd := cli.NewRootCmdWithFS(memFS, execRunner)
	rootCmd.SetOut(dryBuf)
	rootCmd.SetArgs([]string{"upgrade", "/legacy", "--dry-run"})

	err := rootCmd.ExecuteContext(context.Background())
	if err != nil {
		t.Fatalf("upgrade --dry-run failed: %v", err)
	}

	dryOut := dryBuf.String()
	if !strings.Contains(dryOut, "[Dry Run]") || !strings.Contains(dryOut, "restore_region") {
		t.Errorf("expected dry run plan in output:\n%s", dryOut)
	}

	// Verify files were not modified in dry run
	mBytes, _ := memFS.ReadFile("/legacy/loy.yaml")
	if strings.Contains(string(mBytes), "version: 1") {
		t.Errorf("dry run should not modify loy.yaml")
	}

	// 2. Test Apply
	applyBuf := new(bytes.Buffer)
	applyCmd := cli.NewRootCmdWithFS(memFS, execRunner)
	applyCmd.SetOut(applyBuf)
	applyCmd.SetArgs([]string{"upgrade", "/legacy"})

	err = applyCmd.ExecuteContext(context.Background())
	if err != nil {
		t.Fatalf("upgrade apply failed: %v", err)
	}

	applyOut := applyBuf.String()
	if !strings.Contains(applyOut, "Successfully applied") {
		t.Errorf("expected successfully applied message, got: %s", applyOut)
	}

	// Verify loy.yaml upgraded
	updatedMBytes, _ := memFS.ReadFile("/legacy/loy.yaml")
	if !strings.Contains(string(updatedMBytes), "version: 1") {
		t.Errorf("expected version: 1 in loy.yaml:\n%s", string(updatedMBytes))
	}
	if !strings.Contains(string(updatedMBytes), "defaults:") {
		t.Errorf("expected defaults: in loy.yaml:\n%s", string(updatedMBytes))
	}

	// Verify wiring.go has restored region
	updatedWiringBytes, _ := memFS.ReadFile("/legacy/internal/app/wiring.go")
	if !strings.Contains(string(updatedWiringBytes), "// loy:region:routes") {
		t.Errorf("expected // loy:region:routes restored in wiring.go:\n%s", string(updatedWiringBytes))
	}
}

func TestUpgradeCommand_JSONOutput(t *testing.T) {
	memFS := filesystem.NewMemFileSystem()
	execRunner := process.NewExecRunner()

	_ = memFS.WriteFile("/jproj/loy.yaml", []byte(`project:
  name: jproj
`), 0644)

	buf := new(bytes.Buffer)
	rootCmd := cli.NewRootCmdWithFS(memFS, execRunner)
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"upgrade", "/jproj", "--dry-run", "--json"})

	err := rootCmd.ExecuteContext(context.Background())
	if err != nil {
		t.Fatalf("upgrade --json failed: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, `"status":"dry_run"`) || !strings.Contains(out, `"actions":`) {
		t.Errorf("expected JSON upgrade output, got: %s", out)
	}
}
