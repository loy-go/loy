package plan_test

import (
	"context"
	"testing"

	"github.com/loy-go/loy/internal/filesystem"
	"github.com/loy-go/loy/internal/generator"
	"github.com/loy-go/loy/internal/generator/model"
	"github.com/loy-go/loy/internal/generator/plan"
)

func TestPlan_CoverageEdgeCases(t *testing.T) {
	ctx := context.Background()

	t.Run("skip operation ignored in execution", func(t *testing.T) {
		fs := filesystem.NewMemFileSystem()
		p := &plan.Plan{
			Operations: []plan.Operation{
				{
					Type: plan.OpSkip,
					Path: "ignored.txt",
				},
			},
		}
		executor := plan.NewExecutor(fs)
		if err := executor.Execute(ctx, p); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		exists, _ := fs.Exists("ignored.txt")
		if exists {
			t.Error("OpSkip should not create file")
		}
	})

	t.Run("deep directory creation on write", func(t *testing.T) {
		fs := filesystem.NewMemFileSystem()
		p := &plan.Plan{
			TargetDirectory: "root/sub",
			Operations: []plan.Operation{
				{
					Type:        plan.OpCreate,
					Path:        "deep/nested/file.txt",
					Content:     []byte("hello"),
					Permissions: 0755,
				},
			},
		}
		executor := plan.NewExecutor(fs)
		if err := executor.Execute(ctx, p); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		data, err := fs.ReadFile("root/sub/deep/nested/file.txt")
		if err != nil || string(data) != "hello" {
			t.Errorf("expected deep nested file with hello, got %q (err: %v)", string(data), err)
		}
	})

	t.Run("default conflict branch when not developer or generated", func(t *testing.T) {
		fs := filesystem.NewMemFileSystem()
		_ = fs.WriteFile("custom.txt", []byte("existing"), 0644)
		detector := plan.NewConflictDetector(fs)

		art := model.Artifact{
			Path:      "custom.txt",
			Content:   []byte("differing"),
			Ownership: "unknown_custom_ownership",
		}

		_, err := detector.Check(ctx, "", art, generator.Options{Force: false})
		if err == nil {
			t.Fatal("expected conflict error for unknown ownership collision")
		}

		op, err := detector.Check(ctx, "", art, generator.Options{Force: true})
		if err != nil || op != plan.OpOverwrite {
			t.Fatalf("expected overwrite with force, got %s (err: %v)", op, err)
		}
	})
}
