package builtin

import (
	"regexp"
	"strings"
)

// ValidIdentifierRegex enforces alphanumeric/underscore naming rules per AGENTS.md §2.5.
var ValidIdentifierRegex = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)

func splitRawArgs(s string) []string {
	parts := strings.Fields(s)
	var res []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			res = append(res, p)
		}
	}
	return res
}
