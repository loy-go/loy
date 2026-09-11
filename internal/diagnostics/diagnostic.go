package diagnostics

import "fmt"

// Diagnostic models structured diagnostic reports per Doc 14.
type Diagnostic struct {
	Severity Severity `json:"severity"`
	Code     string   `json:"code"`
	Message  string   `json:"message"`
	Detail   string   `json:"detail,omitempty"`
	Hint     string   `json:"hint,omitempty"`
	File     string   `json:"file,omitempty"`
	Line     int      `json:"line,omitempty"`
	Column   int      `json:"column,omitempty"`
}

func (d Diagnostic) Error() string {
	if d.File != "" && d.Line > 0 {
		return fmt.Sprintf("[%s] %s:%d: %s", d.Code, d.File, d.Line, d.Message)
	}
	return fmt.Sprintf("[%s] %s", d.Code, d.Message)
}

// NewError creates an error-severity diagnostic.
func NewError(code, message string) Diagnostic {
	return Diagnostic{
		Severity: SeverityError,
		Code:     code,
		Message:  message,
	}
}

// NewWarning creates a warning-severity diagnostic.
func NewWarning(code, message string) Diagnostic {
	return Diagnostic{
		Severity: SeverityWarning,
		Code:     code,
		Message:  message,
	}
}

// NewInfo creates an info-severity diagnostic.
func NewInfo(code, message string) Diagnostic {
	return Diagnostic{
		Severity: SeverityInfo,
		Code:     code,
		Message:  message,
	}
}
