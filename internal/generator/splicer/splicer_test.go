package splicer_test

import (
	"strings"
	"testing"

	"github.com/loy-go/loy/internal/diagnostics"
	"github.com/loy-go/loy/internal/generator/splicer"
)

func TestSplicer_SpliceRegion_ValidGo(t *testing.T) {
	initial := `package app

func RegisterServices() {
	// loy:region:services
	// loy:endregion
}
`
	sp := splicer.New()
	entry := `registerAccountService()`

	res, err := sp.SpliceRegion([]byte(initial), "services", entry, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	resultStr := string(res)
	if !strings.Contains(resultStr, "registerAccountService()") {
		t.Errorf("entry not found in result: %s", resultStr)
	}

	// Splicing a second time should be idempotent
	res2, err := sp.SpliceRegion(res, "services", entry, true)
	if err != nil {
		t.Fatalf("second splice failed: %v", err)
	}

	if string(res2) != string(res) {
		t.Errorf("expected idempotent result, got diff:\nBefore:\n%s\nAfter:\n%s", string(res), string(res2))
	}
}

func TestSplicer_SpliceRegion_MultiLineIdempotency(t *testing.T) {
	initial := `package app

func RegisterServices() {
	// loy:region:services
	// loy:endregion
}
`
	sp := splicer.New()
	multiLineEntry := `repo := NewUserRepository()
svc := NewUserService(repo)`

	res, err := sp.SpliceRegion([]byte(initial), "services", multiLineEntry, true)
	if err != nil {
		t.Fatalf("first multi-line splice failed: %v", err)
	}

	res2, err := sp.SpliceRegion(res, "services", multiLineEntry, true)
	if err != nil {
		t.Fatalf("second multi-line splice failed: %v", err)
	}

	if string(res2) != string(res) {
		t.Errorf("multi-line splice was not idempotent:\nFirst:\n%s\nSecond:\n%s", string(res), string(res2))
	}
}

func TestSplicer_SpliceRegion_NonGo(t *testing.T) {
	initial := `# config
services:
  # loy:region:services
  # loy:endregion
`
	sp := splicer.New()
	entry := "  - name: auth"

	res, err := sp.SpliceRegion([]byte(initial), "services", entry, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(string(res), "- name: auth") {
		t.Errorf("entry missing in non-go output: %s", string(res))
	}
}

func TestSplicer_CorruptRegions(t *testing.T) {
	sp := splicer.New()

	t.Run("missing target region", func(t *testing.T) {
		content := `package test
// loy:region:other
// loy:endregion
`
		_, err := sp.SpliceRegion([]byte(content), "missing", "foo()", true)
		if err == nil {
			t.Fatal("expected error for missing region, got nil")
		}
		diag, ok := err.(*diagnostics.Diagnostic)
		if !ok || diag.Code != diagnostics.CodeGenRegionCorrupt {
			t.Errorf("expected CodeGenRegionCorrupt, got %v", err)
		}
	})

	t.Run("nested regions", func(t *testing.T) {
		content := `package test
// loy:region:outer
// loy:region:inner
// loy:endregion
// loy:endregion
`
		_, err := sp.SpliceRegion([]byte(content), "outer", "foo()", true)
		if err == nil {
			t.Fatal("expected error for nested regions, got nil")
		}
		diag, ok := err.(*diagnostics.Diagnostic)
		if !ok || diag.Code != diagnostics.CodeGenRegionCorrupt {
			t.Errorf("expected CodeGenRegionCorrupt, got %v", err)
		}
	})

	t.Run("unclosed region", func(t *testing.T) {
		content := `package test
// loy:region:unclosed
func foo() {}
`
		_, err := sp.SpliceRegion([]byte(content), "unclosed", "foo()", true)
		if err == nil {
			t.Fatal("expected error for unclosed region, got nil")
		}
		diag, ok := err.(*diagnostics.Diagnostic)
		if !ok || diag.Code != diagnostics.CodeGenRegionCorrupt {
			t.Errorf("expected CodeGenRegionCorrupt, got %v", err)
		}
	})

	t.Run("unexpected endregion", func(t *testing.T) {
		content := `package test
// loy:endregion
`
		_, err := sp.SpliceRegion([]byte(content), "foo", "foo()", true)
		if err == nil {
			t.Fatal("expected error for unexpected endregion, got nil")
		}
		diag, ok := err.(*diagnostics.Diagnostic)
		if !ok || diag.Code != diagnostics.CodeGenRegionCorrupt {
			t.Errorf("expected CodeGenRegionCorrupt, got %v", err)
		}
	})

	t.Run("duplicate region name", func(t *testing.T) {
		content := `package test
// loy:region:dup
// loy:endregion
// loy:region:dup
// loy:endregion
`
		_, err := sp.SpliceRegion([]byte(content), "dup", "foo()", true)
		if err == nil {
			t.Fatal("expected error for duplicate region name, got nil")
		}
		diag, ok := err.(*diagnostics.Diagnostic)
		if !ok || diag.Code != diagnostics.CodeGenRegionCorrupt {
			t.Errorf("expected CodeGenRegionCorrupt, got %v", err)
		}
	})

	t.Run("empty region name", func(t *testing.T) {
		content := `package test
// loy:region:
// loy:endregion
`
		_, err := sp.SpliceRegion([]byte(content), "any", "foo()", true)
		if err == nil {
			t.Fatal("expected error for empty region name, got nil")
		}
		diag, ok := err.(*diagnostics.Diagnostic)
		if !ok || diag.Code != diagnostics.CodeGenRegionCorrupt {
			t.Errorf("expected CodeGenRegionCorrupt, got %v", err)
		}
	})
}

func TestSplicer_GofmtFailureAfterSplice(t *testing.T) {
	sp := splicer.New()
	initial := `package app

func Register() {
	// loy:region:services
	// loy:endregion
}
`
	// Splicing broken syntax into Go file
	entry := `broken syntax !!! {{{`
	_, err := sp.SpliceRegion([]byte(initial), "services", entry, true)
	if err == nil {
		t.Fatal("expected gofmt failure diagnostic, got nil")
	}

	diag, ok := err.(*diagnostics.Diagnostic)
	if !ok || diag.Code != diagnostics.CodeGenTemplateError {
		t.Errorf("expected CodeGenTemplateError, got %v", err)
	}
}
