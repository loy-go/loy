package diagnostics

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

// ANSI color codes
const (
	ansiReset  = "\033[0m"
	ansiRed    = "\033[31m"
	ansiYellow = "\033[33m"
	ansiCyan   = "\033[36m"
	ansiGray   = "\033[90m"
)

// Formatter formats diagnostics for human or machine consumption.
type Formatter interface {
	Format(w io.Writer, diags []Diagnostic) error
}

// GitHubWorkflowFormatter renders GitHub Actions workflow commands (annotations).
type GitHubWorkflowFormatter struct{}

func escapeWorkflowData(s string) string {
	s = strings.ReplaceAll(s, "%", "%25")
	s = strings.ReplaceAll(s, "\r", "%0D")
	return strings.ReplaceAll(s, "\n", "%0A")
}

func escapeWorkflowProperty(s string) string {
	s = escapeWorkflowData(s)
	s = strings.ReplaceAll(s, ":", "%3A")
	return strings.ReplaceAll(s, ",", "%2C")
}

func (f *GitHubWorkflowFormatter) Format(w io.Writer, diags []Diagnostic) error {
	for _, d := range diags {
		cmdType := "error"
		switch d.Severity {
		case SeverityWarning:
			cmdType = "warning"
		case SeverityInfo:
			cmdType = "notice"
		}

		var params []string
		if d.File != "" {
			params = append(params, fmt.Sprintf("file=%s", escapeWorkflowProperty(d.File)))
		}
		if d.Line > 0 {
			params = append(params, fmt.Sprintf("line=%d", d.Line))
		}
		if d.Code != "" {
			params = append(params, fmt.Sprintf("title=%s", escapeWorkflowProperty(d.Code)))
		}

		paramStr := ""
		if len(params) > 0 {
			paramStr = " " + strings.Join(params, ",")
		}

		msg := d.Message
		if d.Hint != "" {
			msg += fmt.Sprintf(" (Hint: %s)", d.Hint)
		}

		if _, err := fmt.Fprintf(w, "::%s%s::%s\n", cmdType, paramStr, escapeWorkflowData(msg)); err != nil {
			return err
		}
	}
	return nil
}

// JSONFormatter serializes diagnostics to JSON.
type JSONFormatter struct {
	Indent bool
}

func (f *JSONFormatter) Format(w io.Writer, diags []Diagnostic) error {
	enc := json.NewEncoder(w)
	if f.Indent {
		enc.SetIndent("", "  ")
	}
	if diags == nil {
		diags = []Diagnostic{}
	}
	return enc.Encode(diags)
}

// HumanFormatter renders readable console diagnostics.
type HumanFormatter struct {
	Color bool
}

func (f *HumanFormatter) Format(w io.Writer, diags []Diagnostic) error {
	for _, d := range diags {
		var prefix string
		switch d.Severity {
		case SeverityError:
			prefix = "ERROR"
			if f.Color {
				prefix = ansiRed + prefix + ansiReset
			}
		case SeverityWarning:
			prefix = "WARN "
			if f.Color {
				prefix = ansiYellow + prefix + ansiReset
			}
		case SeverityInfo:
			prefix = "INFO "
			if f.Color {
				prefix = ansiCyan + prefix + ansiReset
			}
		}

		codeStr := d.Code
		if f.Color {
			codeStr = ansiGray + "[" + codeStr + "]" + ansiReset
		} else {
			codeStr = "[" + codeStr + "]"
		}

		loc := ""
		if d.File != "" {
			if d.Line > 0 {
				loc = fmt.Sprintf(" %s:%d", d.File, d.Line)
			} else {
				loc = " " + d.File
			}
		}

		if _, err := fmt.Fprintf(w, "%s %s%s: %s\n", prefix, codeStr, loc, d.Message); err != nil {
			return err
		}

		if d.Detail != "" {
			if _, err := fmt.Fprintf(w, "   Detail: %s\n", d.Detail); err != nil {
				return err
			}
		}
		if d.Hint != "" {
			hintPrefix := "   Hint:"
			if f.Color {
				hintPrefix = ansiCyan + hintPrefix + ansiReset
			}
			if _, err := fmt.Fprintf(w, "%s %s\n", hintPrefix, d.Hint); err != nil {
				return err
			}
		}
	}
	return nil
}

// HasErrors returns true if any diagnostic in the slice is SeverityError.
func HasErrors(diags []Diagnostic) bool {
	for _, d := range diags {
		if strings.EqualFold(string(d.Severity), string(SeverityError)) {
			return true
		}
	}
	return false
}
