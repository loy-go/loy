package architecture

import (
	"fmt"

	"github.com/uloydev/loy/internal/diagnostics"
)

// Violation describes a specific architecture rule failure.
type Violation struct {
	RuleID      string
	Code        string
	Message     string
	Detail      string
	Hint        string
	File        string
	Line        int
	Column      int
	Suppressible bool
}

// ToDiagnostic converts a Violation into a standard loyalty Diagnostic.
func (v Violation) ToDiagnostic() diagnostics.Diagnostic {
	code := v.Code
	if code == "" {
		code = fmt.Sprintf("LOY-%s", v.RuleID)
	}
	return diagnostics.Diagnostic{
		Severity: diagnostics.SeverityError,
		Code:     code,
		Message:  v.Message,
		Detail:   v.Detail,
		Hint:     v.Hint,
		File:     v.File,
		Line:     v.Line,
		Column:   v.Column,
	}
}
