package plan_test

import (
	"context"
	"strings"
	"testing"

	"github.com/uloydev/loy/internal/diagnostics"
	"github.com/uloydev/loy/internal/filesystem"
	"github.com/uloydev/loy/internal/generator"
	"github.com/uloydev/loy/internal/generator/model"
	"github.com/uloydev/loy/internal/generator/plan"
)

func TestPlan_ConflictDetection(t *testing.T) {
	ctx := context.Background()

	t.Run("OpCreate when file not exist", func(t *testing.T) {
		fs := filesystem.NewMemFileSystem()
		detector := plan.NewConflictDetector(fs)

		art := model.Artifact{
			Path:      "internal/service/user.go",
			Content:   []byte("package service"),
			Ownership: model.DeveloperOwned,
		}

		op, err := detector.Check(ctx, "app", art, generator.Options{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if op != plan.OpCreate {
			t.Errorf("got op %s, want %s", op, plan.OpCreate)
		}
	})

	t.Run("OpSkip when identical content exists", func(t *testing.T) {
		fs := filesystem.NewMemFileSystem()
		_ = fs.WriteFile("app/user.go", []byte("package service"), 0644)
		detector := plan.NewConflictDetector(fs)

		art := model.Artifact{
			Path:      "user.go",
			Content:   []byte("package service"),
			Ownership: model.DeveloperOwned,
		}

		op, err := detector.Check(ctx, "app", art, generator.Options{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if op != plan.OpSkip {
			t.Errorf("got op %s, want %s", op, plan.OpSkip)
		}
	})

	t.Run("DeveloperOwned collision fails without force", func(t *testing.T) {
		fs := filesystem.NewMemFileSystem()
		_ = fs.WriteFile("user.go", []byte("package old"), 0644)
		detector := plan.NewConflictDetector(fs)

		art := model.Artifact{
			Path:      "user.go",
			Content:   []byte("package new"),
			Ownership: model.DeveloperOwned,
		}

		_, err := detector.Check(ctx, "", art, generator.Options{Force: false})
		if err == nil {
			t.Fatal("expected conflict error, got nil")
		}
		diag, ok := err.(*diagnostics.Diagnostic)
		if !ok || diag.Code != diagnostics.CodeGenConflict {
			t.Errorf("expected CodeGenConflict diagnostic, got %v", err)
		}
	})

	t.Run("DeveloperOwned overwrite succeeds with force", func(t *testing.T) {
		fs := filesystem.NewMemFileSystem()
		_ = fs.WriteFile("user.go", []byte("package old"), 0644)
		detector := plan.NewConflictDetector(fs)

		art := model.Artifact{
			Path:      "user.go",
			Content:   []byte("package new"),
			Ownership: model.DeveloperOwned,
		}

		op, err := detector.Check(ctx, "", art, generator.Options{Force: true})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if op != plan.OpOverwrite {
			t.Errorf("got op %s, want %s", op, plan.OpOverwrite)
		}
	})

	t.Run("GeneratedOwned with intact header overwrites cleanly", func(t *testing.T) {
		fs := filesystem.NewMemFileSystem()
		oldContent := model.GeneratedFileHeader + "\npackage service\n"
		_ = fs.WriteFile("gen.go", []byte(oldContent), 0644)
		detector := plan.NewConflictDetector(fs)

		art := model.Artifact{
			Path:      "gen.go",
			Content:   []byte(model.GeneratedFileHeader + "\npackage service2\n"),
			Ownership: model.GeneratedOwned,
		}

		op, err := detector.Check(ctx, "", art, generator.Options{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if op != plan.OpOverwrite {
			t.Errorf("got op %s, want %s", op, plan.OpOverwrite)
		}
	})

	t.Run("GeneratedOwned modified by developer fails without force", func(t *testing.T) {
		fs := filesystem.NewMemFileSystem()
		_ = fs.WriteFile("gen.go", []byte("package custom_code_without_header"), 0644)
		detector := plan.NewConflictDetector(fs)

		art := model.Artifact{
			Path:      "gen.go",
			Content:   []byte(model.GeneratedFileHeader + "\npackage new\n"),
			Ownership: model.GeneratedOwned,
		}

		_, err := detector.Check(ctx, "", art, generator.Options{Force: false})
		if err == nil {
			t.Fatal("expected conflict error, got nil")
		}
		diag, ok := err.(*diagnostics.Diagnostic)
		if !ok || diag.Code != diagnostics.CodeGenConflict {
			t.Errorf("expected CodeGenConflict, got %v", err)
		}
	})

	t.Run("MixedOwned returns OpSplice", func(t *testing.T) {
		fs := filesystem.NewMemFileSystem()
		_ = fs.WriteFile("wiring.go", []byte("package app"), 0644)
		detector := plan.NewConflictDetector(fs)

		art := model.Artifact{
			Path:      "wiring.go",
			Content:   []byte("registerService()"),
			Ownership: model.MixedOwned,
			Region:    "services",
		}

		op, err := detector.Check(ctx, "", art, generator.Options{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if op != plan.OpSplice {
			t.Errorf("got op %s, want %s", op, plan.OpSplice)
		}
	})
}

func TestPlan_BuilderAndAtomicExecution(t *testing.T) {
	ctx := context.Background()
	fs := filesystem.NewMemFileSystem()

	// Setup existing wiring file for splicing
	initialWiring := `package app

func Wire() {
	// loy:region:services
	// loy:endregion
}
`
	if err := fs.WriteFile("app/wiring.go", []byte(initialWiring), 0644); err != nil {
		t.Fatalf("setup write failed: %v", err)
	}

	builder := plan.NewBuilder(fs)
	artifacts := []model.Artifact{
		{
			Path:      "internal/service/account.go",
			Content:   []byte("package service\n"),
			Ownership: model.DeveloperOwned,
		},
		{
			Path:      "app/wiring.go",
			Content:   []byte("initAccount()"),
			Ownership: model.MixedOwned,
			Region:    "services",
		},
	}

	p, err := builder.Build(ctx, "", artifacts, generator.Options{})
	if err != nil {
		t.Fatalf("build plan failed: %v", err)
	}

	if len(p.Operations) != 2 {
		t.Fatalf("got %d operations, want 2", len(p.Operations))
	}
	if p.Operations[0].Type != plan.OpCreate {
		t.Errorf("first op = %s, want %s", p.Operations[0].Type, plan.OpCreate)
	}
	if p.Operations[1].Type != plan.OpSplice {
		t.Errorf("second op = %s, want %s", p.Operations[1].Type, plan.OpSplice)
	}

	executor := plan.NewExecutor(fs)
	if err := executor.Execute(ctx, p); err != nil {
		t.Fatalf("execute plan failed: %v", err)
	}

	// Verify service file was written
	svcExists, _ := fs.Exists("internal/service/account.go")
	if !svcExists {
		t.Fatal("expected internal/service/account.go to exist")
	}

	// Verify wiring file was spliced
	wiringBytes, _ := fs.ReadFile("app/wiring.go")
	if !strings.Contains(string(wiringBytes), "initAccount()") {
		t.Errorf("expected spliced entry in wiring file: %s", string(wiringBytes))
	}
}

func TestPlan_AtomicRollbackOnFailure(t *testing.T) {
	ctx := context.Background()
	fs := filesystem.NewMemFileSystem()

	origFile := "pre_existing.txt"
	origContent := []byte("original pre-existing data")
	_ = fs.WriteFile(origFile, origContent, 0644)

	// Construct plan where second op fails because region does not exist
	p := &plan.Plan{
		Operations: []plan.Operation{
			{
				Type:        plan.OpCreate,
				Path:        "newly_created.txt",
				Content:     []byte("temp content"),
				Permissions: 0644,
			},
			{
				Type:        plan.OpSplice,
				Path:        origFile,
				Content:     []byte("broken entry"),
				Region:      "non_existent_region", // will fail in splicer
				Permissions: 0644,
			},
		},
	}

	executor := plan.NewExecutor(fs)
	err := executor.Execute(ctx, p)
	if err == nil {
		t.Fatal("expected execution error, got nil")
	}

	// Verify newly_created.txt was rolled back (deleted)
	createdExists, _ := fs.Exists("newly_created.txt")
	if createdExists {
		t.Error("expected newly_created.txt to be removed after rollback")
	}

	// Verify origFile was preserved
	restoredContent, _ := fs.ReadFile(origFile)
	if string(restoredContent) != string(origContent) {
		t.Errorf("restored content mismatch: got %q, want %q", string(restoredContent), string(origContent))
	}
}
