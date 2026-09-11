package renderer_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/uloydev/loy/internal/diagnostics"
	"github.com/uloydev/loy/internal/generator/renderer"
)

type sampleViewModel struct {
	EntityName string
	Fields     []string
}

func TestRenderer_Render_Text(t *testing.T) {
	r := renderer.NewTextRenderer(nil)
	tmpl := `name: {{ snake .EntityName }}
plural: {{ plural (snake .EntityName) }}
pkg: {{ pkg .EntityName }}
`
	model := sampleViewModel{EntityName: "UserAccount"}
	got, err := r.Render(context.Background(), "test.yaml", tmpl, model)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "name: user_account\nplural: user_accounts\npkg: useraccount\n"
	if string(got) != expected {
		t.Errorf("got %q, want %q", string(got), expected)
	}
}

func TestRenderer_RenderGo_Valid(t *testing.T) {
	r := renderer.NewTextRenderer(nil)
	tmpl := `package {{ pkg .EntityName }}

// {{ pascal .EntityName }}Service handles business operations.
type {{ pascal .EntityName }}Service struct {
	id string
}

func New{{ pascal .EntityName }}Service(id string) *{{ pascal .EntityName }}Service {
	return &{{ pascal .EntityName }}Service{id: id}
}
`
	model := sampleViewModel{EntityName: "UserAccount"}
	got, err := r.RenderGo(context.Background(), "service.go", tmpl, model)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify gofmt formatted output
	if !strings.Contains(string(got), "package useraccount") {
		t.Errorf("expected package useraccount in formatted output: %s", string(got))
	}
}

func TestRenderer_RenderGo_GoldenComparison(t *testing.T) {
	r := renderer.NewTextRenderer(nil)
	tmpl := `package service

import "context"

// {{ pascal .EntityName }}Repository defines repository contract.
type {{ pascal .EntityName }}Repository interface {
	FindByID(ctx context.Context, id string) (*{{ pascal .EntityName }}, error)
}

// {{ pascal .EntityName }} entity model.
type {{ pascal .EntityName }} struct {
	ID string
}
`
	model := sampleViewModel{EntityName: "account"}
	got, err := r.RenderGo(context.Background(), "service.go", tmpl, model)
	if err != nil {
		t.Fatalf("RenderGo failed: %v", err)
	}

	goldenPath := filepath.Join("testdata", "golden", "service.go")
	goldenBytes, err := os.ReadFile(goldenPath)
	if err != nil {
		t.Fatalf("reading golden file: %v", err)
	}

	if string(got) != string(goldenBytes) {
		t.Errorf("Golden mismatch!\nGot:\n%s\nWant:\n%s", string(got), string(goldenBytes))
	}
}

func TestRenderer_RenderGo_SyntaxErrorDiagnostic(t *testing.T) {
	r := renderer.NewTextRenderer(nil)
	// Invalid Go source (missing closing brace)
	tmpl := `package test
func Broken( {
`
	_, err := r.RenderGo(context.Background(), "broken.go", tmpl, nil)
	if err == nil {
		t.Fatal("expected syntax error, got nil")
	}

	diag, ok := err.(*diagnostics.Diagnostic)
	if !ok {
		t.Fatalf("expected *diagnostics.Diagnostic, got %T", err)
	}

	if diag.Code != diagnostics.CodeGenTemplateError {
		t.Errorf("got code %s, want %s", diag.Code, diagnostics.CodeGenTemplateError)
	}
	if !strings.Contains(diag.Detail, "Rendered source:") {
		t.Errorf("diagnostic detail missing rendered source line numbers: %s", diag.Detail)
	}
}

func TestRenderer_Render_TemplateParseError(t *testing.T) {
	r := renderer.NewTextRenderer(nil)
	tmpl := `{{ unclosed`
	_, err := r.Render(context.Background(), "invalid.tmpl", tmpl, nil)
	if err == nil {
		t.Fatal("expected parse error, got nil")
	}

	diag, ok := err.(*diagnostics.Diagnostic)
	if !ok {
		t.Fatalf("expected *diagnostics.Diagnostic, got %T", err)
	}
	if diag.Code != diagnostics.CodeGenTemplateError {
		t.Errorf("got code %s, want %s", diag.Code, diagnostics.CodeGenTemplateError)
	}
}

func TestRenderer_Render_TemplateExecuteError(t *testing.T) {
	r := renderer.NewTextRenderer(nil)
	tmpl := `{{ .MissingMethod }}`
	type dummy struct{}
	_, err := r.Render(context.Background(), "invalid.tmpl", tmpl, dummy{})
	if err == nil {
		t.Fatal("expected execution error, got nil")
	}

	diag, ok := err.(*diagnostics.Diagnostic)
	if !ok {
		t.Fatalf("expected *diagnostics.Diagnostic, got %T", err)
	}
	if diag.Code != diagnostics.CodeGenTemplateError {
		t.Errorf("got code %s, want %s", diag.Code, diagnostics.CodeGenTemplateError)
	}
}

func TestRenderer_RenderGo_UnderlyingRenderError(t *testing.T) {
	r := renderer.NewTextRenderer(nil)
	tmpl := `{{ .BrokenField }}`
	_, err := r.RenderGo(context.Background(), "invalid.go", tmpl, struct{}{})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
