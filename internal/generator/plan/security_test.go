package plan_test

import (
	"context"
	"testing"

	"github.com/loy-go/loy/internal/diagnostics"
	"github.com/loy-go/loy/internal/filesystem"
	"github.com/loy-go/loy/internal/generator"
	"github.com/loy-go/loy/internal/generator/model"
	"github.com/loy-go/loy/internal/generator/plan"
)

func TestPlan_PathTraversalSecurity(t *testing.T) {
	ctx := context.Background()
	fs := filesystem.NewMemFileSystem()

	t.Run("conflict detector rejects path traversal", func(t *testing.T) {
		detector := plan.NewConflictDetector(fs)
		art := model.Artifact{
			Path:      "../../etc/passwd",
			Content:   []byte("root:x"),
			Ownership: model.DeveloperOwned,
		}

		_, err := detector.Check(ctx, "subapp", art, generator.Options{})
		if err == nil {
			t.Fatal("expected path traversal error, got nil")
		}

		diag, ok := err.(*diagnostics.Diagnostic)
		if !ok || diag.Code != diagnostics.CodeFSPathTraversal {
			t.Errorf("expected CodeFSPathTraversal diagnostic, got %v", err)
		}
	})

	t.Run("executor rejects path traversal and rolls back", func(t *testing.T) {
		executor := plan.NewExecutor(fs)
		p := &plan.Plan{
			TargetDirectory: "subapp",
			Operations: []plan.Operation{
				{
					Type:        plan.OpCreate,
					Path:        "../../../escaped.txt",
					Content:     []byte("danger"),
					Permissions: 0644,
				},
			},
		}

		err := executor.Execute(ctx, p)
		if err == nil {
			t.Fatal("expected executor to reject path traversal, got nil")
		}

		diag, ok := err.(*diagnostics.Diagnostic)
		if !ok || diag.Code != diagnostics.CodeFSPathTraversal {
			t.Errorf("expected CodeFSPathTraversal diagnostic, got %v", err)
		}
	})
}
