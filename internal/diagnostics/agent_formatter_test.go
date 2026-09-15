package diagnostics_test

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/loy-go/loy/internal/diagnostics"
)

func TestAgentPromptFormatter(t *testing.T) {
	t.Run("empty diagnostics formats empty json array", func(t *testing.T) {
		formatter := &diagnostics.AgentPromptFormatter{}
		var buf bytes.Buffer
		err := formatter.Format(&buf, []diagnostics.Diagnostic{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		expected := "[]\n"
		if buf.String() != expected {
			t.Errorf("expected %q, got %q", expected, buf.String())
		}
	})

	t.Run("formats ARCH-002 domain to infra violation", func(t *testing.T) {
		diag := diagnostics.Diagnostic{
			Severity: diagnostics.SeverityError,
			Code:     "LOY-ARCH-002",
			Message:  "domain package internal/order/domain imports infrastructure package internal/platform/database",
			File:     "internal/order/domain/order.go",
			Line:     14,
		}

		formatter := &diagnostics.AgentPromptFormatter{}
		var buf bytes.Buffer
		err := formatter.Format(&buf, []diagnostics.Diagnostic{diag})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		var agentDiags []diagnostics.AgentDiagnostic
		if err := json.Unmarshal(buf.Bytes(), &agentDiags); err != nil {
			t.Fatalf("failed to unmarshal agent diagnostics JSON: %v\nOutput: %s", err, buf.String())
		}

		if len(agentDiags) != 1 {
			t.Fatalf("expected 1 agent diagnostic, got %d", len(agentDiags))
		}

		ad := agentDiags[0]
		if ad.Code != "LOY-ARCH-002" {
			t.Errorf("expected code LOY-ARCH-002, got %s", ad.Code)
		}
		if ad.File != "internal/order/domain/order.go" || ad.Line != 14 {
			t.Errorf("unexpected file/line: %s:%d", ad.File, ad.Line)
		}
		if ad.Remediation.Action != "invert_dependency" {
			t.Errorf("expected action invert_dependency, got %s", ad.Remediation.Action)
		}
		if !strings.Contains(ad.Remediation.Prompt, "OrderRepository") {
			t.Errorf("expected prompt to suggest repository interface: %s", ad.Remediation.Prompt)
		}
	})

	t.Run("covers all rules ARCH-001 through ARCH-015", func(t *testing.T) {
		ruleCodes := []string{
			"LOY-ARCH-001", "LOY-ARCH-002", "LOY-ARCH-003", "LOY-ARCH-004", "LOY-ARCH-005",
			"LOY-ARCH-006", "LOY-ARCH-007", "LOY-ARCH-008", "LOY-ARCH-009", "LOY-ARCH-010",
			"LOY-ARCH-011", "LOY-ARCH-012", "LOY-ARCH-013", "LOY-ARCH-014", "LOY-ARCH-015",
		}

		for _, code := range ruleCodes {
			diag := diagnostics.Diagnostic{
				Severity: diagnostics.SeverityError,
				Code:     code,
				Message:  "test violation for " + code,
				File:     "internal/test/file.go",
				Line:     10,
			}
			ad := diagnostics.ToAgentDiagnostic(diag)
			if ad.Rationale == "" {
				t.Errorf("missing rationale for %s", code)
			}
			if ad.Remediation.Action == "" || ad.Remediation.Action == "fix_violation" {
				t.Errorf("expected specific remediation action for %s, got %s", code, ad.Remediation.Action)
			}
			if ad.Remediation.Prompt == "" {
				t.Errorf("missing remediation prompt for %s", code)
			}
		}
	})
}
