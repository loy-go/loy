package diagnostics_test

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/loy-go/loy/internal/diagnostics"
)

func TestDiagnosticFormatting_JSON(t *testing.T) {
	diags := []diagnostics.Diagnostic{
		{
			Severity: diagnostics.SeverityError,
			Code:     diagnostics.CodeFSNotFound,
			Message:  "manifest file not found",
			File:     "loy.yaml",
			Line:     1,
			Hint:     "run loy init to create one",
		},
	}

	formatter := &diagnostics.JSONFormatter{Indent: false}
	var buf bytes.Buffer
	err := formatter.Format(&buf, diags)
	if err != nil {
		t.Fatalf("JSON format failed: %v", err)
	}

	var parsed []map[string]any
	if err := json.Unmarshal(buf.Bytes(), &parsed); err != nil {
		t.Fatalf("failed to unmarshal JSON output: %v", err)
	}

	if len(parsed) != 1 {
		t.Fatalf("expected 1 diagnostic in JSON array, got %d", len(parsed))
	}
	if parsed[0]["code"] != diagnostics.CodeFSNotFound {
		t.Fatalf("expected code %s, got %v", diagnostics.CodeFSNotFound, parsed[0]["code"])
	}
}

func TestDiagnosticFormatting_Human(t *testing.T) {
	diags := []diagnostics.Diagnostic{
		{
			Severity: diagnostics.SeverityError,
			Code:     diagnostics.CodeCLIUsageError,
			Message:  "unknown argument provided",
			Detail:   "argument --foo is invalid",
			Hint:     "check available flags with --help",
		},
		{
			Severity: diagnostics.SeverityWarning,
			Code:     "LOY-GEN-100",
			Message:  "file will be overwritten",
		},
	}

	formatter := &diagnostics.HumanFormatter{Color: false}
	var buf bytes.Buffer
	err := formatter.Format(&buf, diags)
	if err != nil {
		t.Fatalf("Human format failed: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "ERROR [LOY-CLI-001]: unknown argument provided") {
		t.Fatalf("expected formatted error line, got: %s", out)
	}
	if !strings.Contains(out, "Detail: argument --foo is invalid") {
		t.Fatalf("expected detail line, got: %s", out)
	}
	if !strings.Contains(out, "Hint: check available flags with --help") {
		t.Fatalf("expected hint line, got: %s", out)
	}
	if !strings.Contains(out, "WARN  [LOY-GEN-100]: file will be overwritten") {
		t.Fatalf("expected warning line, got: %s", out)
	}
}

func TestDiagnosticFormatting_HumanWithColor(t *testing.T) {
	diags := []diagnostics.Diagnostic{
		diagnostics.NewInfo("I01", "starting server"),
		diagnostics.NewWarning("W01", "deprecated flag"),
		diagnostics.NewError("E01", "connection refused"),
	}

	formatter := &diagnostics.HumanFormatter{Color: true}
	var buf bytes.Buffer
	if err := formatter.Format(&buf, diags); err != nil {
		t.Fatal(err)
	}

	out := buf.String()
	if !strings.Contains(out, "INFO") || !strings.Contains(out, "WARN") || !strings.Contains(out, "ERROR") {
		t.Fatalf("expected colored output with severities, got %q", out)
	}
}

func TestDiagnostic_ErrorString(t *testing.T) {
	dWithLoc := diagnostics.Diagnostic{
		Code:    "E01",
		Message: "syntax error",
		File:    "main.go",
		Line:    42,
	}
	if dWithLoc.Error() != "[E01] main.go:42: syntax error" {
		t.Fatalf("unexpected error string: %s", dWithLoc.Error())
	}

	dWithoutLoc := diagnostics.NewError("E02", "general failure")
	if dWithoutLoc.Error() != "[E02] general failure" {
		t.Fatalf("unexpected error string: %s", dWithoutLoc.Error())
	}
}

func TestHasErrors(t *testing.T) {
	onlyWarnings := []diagnostics.Diagnostic{
		diagnostics.NewWarning("W01", "low disk space"),
	}
	if diagnostics.HasErrors(onlyWarnings) {
		t.Fatalf("expected HasErrors to return false for only warnings")
	}

	withError := []diagnostics.Diagnostic{
		diagnostics.NewWarning("W01", "low disk space"),
		diagnostics.NewError("E01", "failed to read"),
	}
	if !diagnostics.HasErrors(withError) {
		t.Fatalf("expected HasErrors to return true")
	}
}
