package architecture_test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/uloydev/loy/internal/architecture"
	"github.com/uloydev/loy/internal/architecture/rules"
	"github.com/uloydev/loy/internal/cli"
	"github.com/uloydev/loy/internal/filesystem"
)

func TestArchitectureCheckOnRuntimeScaffold(t *testing.T) {
	tmpDir := t.TempDir()
	fs := filesystem.NewOSFileSystem()

	// 1. Create a project with api preset
	rootCmd := cli.NewRootCmdWithFS(fs, nil)
	rootCmd.SetArgs([]string{"new", "sampleapp", "--preset", "minimal", "-C", tmpDir})
	err := rootCmd.ExecuteContext(context.Background())
	if err != nil {
		t.Fatalf("loy new failed: %v", err)
	}

	appDir := filepath.Join(tmpDir, "sampleapp")

	// 2. Run architecture analyzer directly over generated files
	analyzer := architecture.NewAnalyzer(fs, architecture.AnalyzerConfig{
		ModuleName: "sampleapp",
		RootDir:    appDir,
		Rules:      rules.DefaultRules(),
	})
	violations, err := analyzer.Run(context.Background())
	if err != nil {
		t.Fatalf("architecture analyze failed: %v", err)
	}

	if len(violations) > 0 {
		for _, v := range violations {
			t.Errorf("unexpected architecture violation in runtime scaffold: [%s] %s at %s:%d", v.RuleID, v.Message, v.File, v.Line)
		}
	}
}
