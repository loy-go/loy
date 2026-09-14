package doctor_test

import (
	"context"
	"strings"
	"testing"

	"github.com/uloydev/loy/internal/diagnostics"
	"github.com/uloydev/loy/internal/doctor"
	"github.com/uloydev/loy/internal/filesystem"
	"github.com/uloydev/loy/internal/process"
)

func TestDoctorRun_Prerequisites(t *testing.T) {
	memFS := filesystem.NewMemFileSystem()
	mockProc := process.NewExecRunner() // real execution for system Go check

	doc := doctor.NewDoctor(memFS, mockProc)
	ctx := context.Background()

	checks, err := doc.Run(ctx, "/testapp")
	if err != nil {
		t.Fatalf("unexpected doctor error: %v", err)
	}

	if len(checks) < 4 {
		t.Fatalf("expected at least 4 check items, got %d", len(checks))
	}

	// Go check should pass in standard dev environment
	var foundGo bool
	for _, c := range checks {
		if c.Name == "Go Compiler" {
			foundGo = true
			if !c.Passed {
				t.Errorf("expected Go compiler check to pass, detail: %s", c.Detail)
			}
			if !strings.Contains(c.Detail, "go") {
				t.Errorf("unexpected Go detail: %s", c.Detail)
			}
		}
	}
	if !foundGo {
		t.Errorf("Go Compiler check missing from results")
	}
}

func TestDoctorRun_ProjectValidation(t *testing.T) {
	memFS := filesystem.NewMemFileSystem()
	mockProc := process.NewExecRunner()

	// 1. In empty directory: Go Module and Loy Manifest checks should fail/warn
	doc := doctor.NewDoctor(memFS, mockProc)
	ctx := context.Background()

	checks, err := doc.Run(ctx, "/empty")
	if err != nil {
		t.Fatalf("unexpected doctor error: %v", err)
	}

	var modWarn, manWarn bool
	for _, c := range checks {
		if c.Name == "Go Module" && !c.Passed {
			modWarn = true
			if c.Diagnostic == nil || c.Diagnostic.Code != diagnostics.CodeDoctorModuleMissing {
				t.Errorf("expected %s diagnostic for missing module, got %v", diagnostics.CodeDoctorModuleMissing, c.Diagnostic)
			}
		}
		if c.Name == "Loy Manifest" && !c.Passed {
			manWarn = true
			if c.Diagnostic == nil || c.Diagnostic.Code != diagnostics.CodeDoctorManifestMissing {
				t.Errorf("expected %s diagnostic for missing manifest, got %v", diagnostics.CodeDoctorManifestMissing, c.Diagnostic)
			}
		}
	}

	if !modWarn || !manWarn {
		t.Errorf("expected warnings for missing module and manifest, got mod: %v, man: %v", modWarn, manWarn)
	}

	// 2. Populate valid go.mod and loy.yaml
	_ = memFS.WriteFile("/valid/go.mod", []byte("module github.com/test/valid\n\ngo 1.22\n"), 0644)
	manifestData := `version: 1
project:
  name: valid
  module: github.com/test/valid
`
	_ = memFS.WriteFile("/valid/loy.yaml", []byte(manifestData), 0644)

	checksValid, err := doc.Run(ctx, "/valid")
	if err != nil {
		t.Fatalf("unexpected doctor error: %v", err)
	}

	for _, c := range checksValid {
		if c.Name == "Go Module" && !c.Passed {
			t.Errorf("expected Go Module to pass in /valid, got: %s", c.Detail)
		}
		if c.Name == "Loy Manifest" && !c.Passed {
			t.Errorf("expected Loy Manifest to pass in /valid, got: %s", c.Detail)
		}
	}
}

func TestDoctorRun_FullstackValidation(t *testing.T) {
	memFS := filesystem.NewMemFileSystem()
	mockProc := process.NewExecRunner()
	doc := doctor.NewDoctor(memFS, mockProc)
	ctx := context.Background()

	_ = memFS.MkdirAll("/fullstack/views", 0755)
	_ = memFS.WriteFile("/fullstack/package.json", []byte(`{"scripts":{"dev":"vite"}}`), 0644)
	_ = memFS.WriteFile("/fullstack/pnpm-lock.yaml", []byte(""), 0644)

	checks, err := doc.Run(ctx, "/fullstack")
	if err != nil {
		t.Fatalf("unexpected doctor error: %v", err)
	}

	var foundTempl, foundNode, foundPnpm bool
	for _, c := range checks {
		if c.Name == "templ CLI" {
			foundTempl = true
		}
		if c.Name == "Node.js" {
			foundNode = true
		}
		if c.Name == "pnpm Package Manager" {
			foundPnpm = true
		}
	}

	if !foundTempl {
		t.Errorf("expected templ CLI check in fullstack directory")
	}
	if !foundNode {
		t.Errorf("expected Node.js check in fullstack directory")
	}
	if !foundPnpm {
		t.Errorf("expected pnpm Package Manager check with pnpm-lock.yaml")
	}
}
