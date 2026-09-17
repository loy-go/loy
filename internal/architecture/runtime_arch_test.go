package architecture_test

import (
	"context"
	"testing"

	"github.com/loy-go/loy/internal/architecture"
	"github.com/loy-go/loy/internal/architecture/rules"
	"github.com/loy-go/loy/internal/cli"
	"github.com/loy-go/loy/internal/filesystem"
	"github.com/loy-go/loy/internal/process"
)

func TestArchitectureCheckOnRuntimeScaffold(t *testing.T) {
	fs := filesystem.NewMemFileSystem()
	runner := process.NewNoopRunner()

	// 1. Create a project with minimal preset in-memory without spawning external processes
	rootCmd := cli.NewRootCmdWithFS(fs, runner)
	rootCmd.SetArgs([]string{"new", "sampleapp", "--preset", "minimal", "--no-tidy"})
	err := rootCmd.ExecuteContext(context.Background())
	if err != nil {
		t.Fatalf("loy new failed: %v", err)
	}

	appDir, _ := filesystem.CleanAndValidatePath(".", "sampleapp")

	// 2. Run architecture analyzer directly over generated files in memory
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
