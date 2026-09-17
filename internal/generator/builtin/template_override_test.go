package builtin_test

import (
	"context"
	"testing"

	"github.com/loy-go/loy/internal/filesystem"
	"github.com/loy-go/loy/internal/generator"
	"github.com/loy-go/loy/internal/generator/builtin"
)

func TestTemplateOverride_Precedence(t *testing.T) {
	memFS := filesystem.NewMemFileSystem()
	customDir := "/project/.loy/templates"
	if err := memFS.MkdirAll(customDir, 0755); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}

	customContent := `// CUSTOM OVERRIDE
package {{.FeaturePkg}}

type {{.Pascal}}Custom struct {
	ID string
}
`
	if err := memFS.WriteFile(customDir+"/model.go.tmpl", []byte(customContent), 0644); err != nil {
		t.Fatalf("write custom template failed: %v", err)
	}

	resolver := builtin.NewFilesystemTemplateResolver(memFS, customDir)
	ctx := builtin.WithTemplateResolver(context.Background(), resolver)

	gen := builtin.NewModelGenerator("github.com/example/app")
	input := generator.Input{
		Name: "order",
		Args: map[string]string{},
	}

	artifacts, err := gen.Generate(ctx, input)
	if err != nil {
		t.Fatalf("generate with override failed: %v", err)
	}
	if len(artifacts) == 0 {
		t.Fatalf("expected artifacts, got 0")
	}

	output := string(artifacts[0].Content)
	if !contains(output, "// CUSTOM OVERRIDE") {
		t.Fatalf("expected custom override in generated output, got:\n%s", output)
	}
	if !contains(output, "OrderCustom") {
		t.Fatalf("expected OrderCustom in generated output, got:\n%s", output)
	}
}

func TestListTemplates(t *testing.T) {
	tmpls, err := builtin.ListTemplates()
	if err != nil {
		t.Fatalf("ListTemplates failed: %v", err)
	}
	if len(tmpls) == 0 {
		t.Fatalf("expected at least one template")
	}

	foundModel := false
	for _, name := range tmpls {
		if name == "model.go.tmpl" {
			foundModel = true
			break
		}
	}
	if !foundModel {
		t.Fatalf("expected model.go.tmpl in templates list")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || func() bool {
		for i := 0; i <= len(s)-len(substr); i++ {
			if s[i:i+len(substr)] == substr {
				return true
			}
		}
		return false
	}())
}
