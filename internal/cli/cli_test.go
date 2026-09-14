package cli_test

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/loy-go/loy/internal/cli"
	"github.com/loy-go/loy/internal/version"
)

func TestVersionCmd_Output(t *testing.T) {
	cmd := cli.NewRootCmd()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"version"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error executing version cmd: %v", err)
	}
}

func TestVersionCmd_JSON(t *testing.T) {
	cmd := cli.NewRootCmd()
	cmd.SetArgs([]string{"--json", "version"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error executing version --json: %v", err)
	}
}

func TestDirectoryFlag(t *testing.T) {
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(origDir)
	})

	cmd := cli.NewRootCmd()
	cmd.SetArgs([]string{"--directory", os.TempDir(), "version"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error with --directory flag: %v", err)
	}
}

func TestVersion_Get(t *testing.T) {
	info := version.Get()
	if info.Version == "" {
		t.Fatal("expected non-empty version")
	}
	if info.GoVersion == "" {
		t.Fatal("expected non-empty go version")
	}

	data, err := json.Marshal(info)
	if err != nil {
		t.Fatalf("failed to marshal version info: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("marshaled version info is empty")
	}
}

func TestInitCmd_Integration(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	_ = os.Chdir(tmpDir)
	t.Cleanup(func() {
		_ = os.Chdir(origDir)
	})

	// Fail if no go.mod
	cmdNoMod := cli.NewRootCmd()
	cmdNoMod.SetArgs([]string{"init", "--preset", "api"})
	if err := cmdNoMod.Execute(); err == nil {
		t.Fatal("expected error running init without go.mod")
	}

	_ = os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte("module testmod\n\ngo 1.22\n"), 0644)

	cmd := cli.NewRootCmd()
	cmd.SetArgs([]string{"init", "--preset", "api"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("init command failed: %v", err)
	}

	manifestPath := filepath.Join(tmpDir, "loy.yaml")
	if _, err := os.Stat(manifestPath); err != nil {
		t.Fatalf("loy.yaml not created: %v", err)
	}

	// Repeated init without --force must fail
	cmdRepeat := cli.NewRootCmd()
	cmdRepeat.SetArgs([]string{"init", "--preset", "api"})
	if err := cmdRepeat.Execute(); err == nil {
		t.Fatal("expected error on re-init without --force")
	}

	// Force overwrite works
	cmdForce := cli.NewRootCmd()
	cmdForce.SetArgs([]string{"init", "--preset", "minimal", "--force"})
	if err := cmdForce.Execute(); err != nil {
		t.Fatalf("expected force overwrite to succeed, got %v", err)
	}

	// Invalid preset returns error
	cmdBadPreset := cli.NewRootCmd()
	cmdBadPreset.SetArgs([]string{"init", "--preset", "invalid", "--force"})
	if err := cmdBadPreset.Execute(); err == nil {
		t.Fatal("expected error with unknown preset")
	}
}

func TestNewCmd_Integration(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	_ = os.Chdir(tmpDir)
	t.Cleanup(func() {
		_ = os.Chdir(origDir)
	})

	cmd := cli.NewRootCmd()
	cmd.SetArgs([]string{"new", "demoapp", "--preset", "fullstack"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("new command failed: %v", err)
	}

	projDir := filepath.Join(tmpDir, "demoapp")
	if _, err := os.Stat(projDir); err != nil {
		t.Fatalf("project dir not created: %v", err)
	}
	if _, err := os.Stat(filepath.Join(projDir, "loy.yaml")); err != nil {
		t.Fatalf("loy.yaml not created in demoapp: %v", err)
	}
	if _, err := os.Stat(filepath.Join(projDir, "go.mod")); err != nil {
		t.Fatalf("go.mod not created in demoapp: %v", err)
	}

	// Duplicate new without --force fails
	cmdDup := cli.NewRootCmd()
	cmdDup.SetArgs([]string{"new", "demoapp"})
	if err := cmdDup.Execute(); err == nil {
		t.Fatal("expected duplicate project creation to fail without --force")
	}

	// Unknown preset fails
	cmdBad := cli.NewRootCmd()
	cmdBad.SetArgs([]string{"new", "badproj", "--preset", "unknown"})
	if err := cmdBad.Execute(); err == nil {
		t.Fatal("expected error with unknown preset in new cmd")
	}

	// Custom adapters & multi-tenant flags
	cmdCustom := cli.NewRootCmd()
	cmdCustom.SetArgs([]string{"new", "customapp", "--http=chi", "--db=sqlite", "--queue=river", "--multi-tenant=rls"})
	if err := cmdCustom.Execute(); err != nil {
		t.Fatalf("new command with custom adapters failed: %v", err)
	}

	customYAML, err := os.ReadFile(filepath.Join(tmpDir, "customapp", "loy.yaml"))
	if err != nil {
		t.Fatalf("failed reading custom loy.yaml: %v", err)
	}
	if !strings.Contains(string(customYAML), "strategy: rls") {
		t.Errorf("expected custom loy.yaml to contain multi-tenancy rls strategy, got:\n%s", string(customYAML))
	}
}

func TestRoutesCmd(t *testing.T) {
	cmd := cli.NewRootCmd()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"routes", "."})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("routes command failed: %v", err)
	}
}

func TestSeedCmd(t *testing.T) {
	tmpDir := t.TempDir()
	_ = os.MkdirAll(filepath.Join(tmpDir, "internal/platform/database/seeds"), 0755)

	cmd := cli.NewRootCmd()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"seed", tmpDir})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("seed command failed: %v", err)
	}
}
