package architecture_test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/uloydev/loy/internal/architecture"
	"github.com/uloydev/loy/internal/architecture/rules"
	"github.com/uloydev/loy/internal/filesystem"
)

func TestAnalyzer_DeepAndCoverage(t *testing.T) {
	memFS := filesystem.NewMemFileSystem()
	_ = memFS.MkdirAll("/proj/internal/domain/user", 0755)
	_ = memFS.WriteFile("/proj/internal/domain/user/user.go", []byte("package user\ntype User struct{}\n"), 0644)
	_ = memFS.WriteFile("/proj/internal/domain/user/user_test.go", []byte("package user\n"), 0644)
	_ = memFS.MkdirAll("/proj/.git", 0755)
	_ = memFS.WriteFile("/proj/.git/config", []byte("git"), 0644)

	cfg := architecture.AnalyzerConfig{
		ModuleName: "github.com/test/proj",
		RootDir:    "/proj",
		Deep:       false,
		Rules:      rules.DefaultRules(),
	}

	analyzer := architecture.NewAnalyzer(memFS, cfg)
	vs, err := analyzer.Run(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(vs) != 0 {
		t.Fatalf("expected 0 violations, got: %v", vs)
	}

	// Test real OS filesystem deep loading fallback without crash
	tmpDir := t.TempDir()
	osFS := filesystem.NewOSFileSystem()
	_ = osFS.WriteFile(filepath.Join(tmpDir, "dummy.go"), []byte("package main\nfunc main(){}\n"), 0644)

	osCfg := architecture.AnalyzerConfig{
		ModuleName: "example.com/test",
		RootDir:    tmpDir,
		Deep:       true,
		Rules:      rules.DefaultRules(),
	}
	osAnalyzer := architecture.NewAnalyzer(osFS, osCfg)
	_, _ = osAnalyzer.Run(context.Background())
}
