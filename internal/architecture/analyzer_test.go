package architecture

import (
	"context"
	"testing"

	"github.com/uloydev/loy/internal/filesystem"
)

// mockRule tests analyzer rule invocation
type mockRule struct {
	id string
	vs []Violation
}

func (m *mockRule) ID() string          { return m.id }
func (m *mockRule) Description() string { return "mock rule" }
func (m *mockRule) Check(ctx context.Context, a *Analysis) []Violation {
	return m.vs
}

func TestAnalyzer(t *testing.T) {
	memFS := filesystem.NewMemFileSystem()
	_ = memFS.MkdirAll("/app/internal/domain/user", 0755)
	_ = memFS.WriteFile("/app/internal/domain/user/user.go", []byte(`package user
// loy:ignore ARCH-002 reason="temporary adapter"
import "github.com/example/app/internal/repository/pg"
type User struct {}
`), 0644)

	_ = memFS.MkdirAll("/app/internal/repository/pg", 0755)
	_ = memFS.WriteFile("/app/internal/repository/pg/pg.go", []byte(`package pg
type Repo struct {}
`), 0644)

	cfg := AnalyzerConfig{
		ModuleName: "github.com/example/app",
		RootDir:    "/app",
		Deep:       false,
		Rules: []Rule{
			&mockRule{
				id: "ARCH-002",
				vs: []Violation{
					{
						RuleID: "ARCH-002",
						File:   "/app/internal/domain/user/user.go",
						Line:   3,
					},
				},
			},
		},
	}

	analyzer := NewAnalyzer(memFS, cfg)
	violations, err := analyzer.Run(context.Background())
	if err != nil {
		t.Fatalf("unexpected analyzer error: %v", err)
	}

	// Should be suppressed because valid loy:ignore exists
	if len(violations) != 0 {
		t.Fatalf("expected 0 violations after suppression, got %d", len(violations))
	}
}
