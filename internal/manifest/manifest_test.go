package manifest_test

import (
	"testing"

	"github.com/loy-go/loy/internal/diagnostics"
	"github.com/loy-go/loy/internal/manifest"
)

func TestParser_ParseStrict_Valid(t *testing.T) {
	data := []byte(`
version: 1
project:
  name: testapp
defaults:
  http: fiber
  database: postgres
`)
	parser := manifest.NewParser()
	m, diag := parser.ParseStrict("loy.yaml", data)
	if diag != nil {
		t.Fatalf("unexpected diagnostic: %+v", diag)
	}
	if m == nil {
		t.Fatal("expected non-nil manifest")
	}
	if m.Version != 1 || m.Project.Name != "testapp" {
		t.Errorf("unexpected manifest content: %+v", m)
	}

	norm := manifest.Normalize(m)
	if norm.Defaults.HTTP != "fiber" || norm.Defaults.Cache != "valkey" {
		t.Errorf("defaults not applied properly: %+v", norm.Defaults)
	}
}

func TestParser_ParseStrict_MultipleDocuments(t *testing.T) {
	data := []byte(`
version: 1
project:
  name: testapp
---
version: 1
project:
  name: testapp2
`)
	parser := manifest.NewParser()
	m, diag := parser.ParseStrict("loy.yaml", data)
	if m != nil {
		t.Fatal("expected nil manifest for multiple documents")
	}
	if diag == nil || diag.Code != diagnostics.CodeConfigSyntaxError {
		t.Fatalf("expected syntax error diagnostic, got %+v", diag)
	}
}

func TestParser_ParseStrict_UnknownField(t *testing.T) {
	data := []byte(`
version: 1
project:
  name: testapp
invalid_field_name: "fail"
`)
	parser := manifest.NewParser()
	m, diag := parser.ParseStrict("loy.yaml", data)
	if m != nil {
		t.Fatalf("expected nil manifest on unknown field, got: %+v", m)
	}
	if diag == nil {
		t.Fatal("expected diagnostic for unknown field")
	}
	if diag.Code != diagnostics.CodeConfigSyntaxError {
		t.Errorf("expected code %s, got %s", diagnostics.CodeConfigSyntaxError, diag.Code)
	}
	if diag.Line != 5 {
		t.Errorf("expected line 5, got %d", diag.Line)
	}
}

func TestParser_ParseStrict_Empty(t *testing.T) {
	parser := manifest.NewParser()
	_, diag := parser.ParseStrict("loy.yaml", []byte("   \n"))
	if diag == nil || diag.Code != diagnostics.CodeConfigSyntaxError {
		t.Fatalf("expected syntax error diagnostic for empty file, got %+v", diag)
	}
}

func TestValidator_Validation(t *testing.T) {
	val := manifest.NewValidator()

	tests := []struct {
		name      string
		manifest  *manifest.Manifest
		wantCodes []string
	}{
		{
			name: "valid manifest",
			manifest: &manifest.Manifest{
				Version: 1,
				Project: manifest.ProjectConfig{Name: "my-app"},
				Defaults: manifest.DefaultsConfig{
					HTTP:     "fiber",
					Database: "postgres",
					Cache:    "valkey",
					Queue:    "asynq",
					Template: "templ",
					Assets:   "vite",
				},
				Integrations: map[string]manifest.Integration{
					"database": {Driver: "sqlite"},
				},
			},
			wantCodes: nil,
		},
		{
			name: "empty name",
			manifest: &manifest.Manifest{
				Version: 1,
				Project: manifest.ProjectConfig{Name: ""},
			},
			wantCodes: []string{diagnostics.CodeConfigValidationError},
		},
		{
			name: "invalid version & invalid name",
			manifest: &manifest.Manifest{
				Version: 2,
				Project: manifest.ProjectConfig{Name: "invalid name with spaces!"},
			},
			wantCodes: []string{diagnostics.CodeConfigValidationError, diagnostics.CodeConfigValidationError},
		},
		{
			name: "unknown drivers and invalid integration capability",
			manifest: &manifest.Manifest{
				Version: 1,
				Project: manifest.ProjectConfig{Name: "app"},
				Defaults: manifest.DefaultsConfig{
					HTTP:     "django",
					Database: "oracle",
				},
				Integrations: map[string]manifest.Integration{
					"unknown_cap": {Driver: "foo"},
				},
			},
			wantCodes: []string{
				diagnostics.CodeConfigValidationError,
				diagnostics.CodeConfigValidationError,
				diagnostics.CodeConfigValidationError,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			diags := val.Validate("loy.yaml", tt.manifest)
			if len(diags) != len(tt.wantCodes) {
				t.Fatalf("expected %d diagnostics, got %d: %+v", len(tt.wantCodes), len(diags), diags)
			}
			for i, code := range tt.wantCodes {
				if diags[i].Code != code {
					t.Errorf("diag[%d]: expected code %s, got %s", i, code, diags[i].Code)
				}
			}
		})
	}
}
