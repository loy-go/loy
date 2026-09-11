package architecture

import (
	"regexp"
	"strings"
)

// Regex pattern: // loy:ignore (ARCH-\d+) reason="(.+)"
var suppressionRegex = regexp.MustCompile(`//\s*loy:ignore\s+(ARCH-\d+)(?:\s+reason="([^"]*)")?`)

// Suppression records an explicit ignore directive in code comments.
type Suppression struct {
	RuleID string
	Reason string
	File   string
	Line   int
	Valid  bool // true if non-empty reason provided
}

// NonSuppressibleRules defines rules that can never be bypassed per Spec 05.
var NonSuppressibleRules = map[string]bool{
	"ARCH-001": true, // cycles
	"ARCH-013": true, // workspace boundary
}

// ParseSuppression extracts a suppression from a raw comment line.
func ParseSuppression(comment string, file string, line int) (Suppression, bool) {
	matches := suppressionRegex.FindStringSubmatch(comment)
	if len(matches) == 0 {
		return Suppression{}, false
	}

	ruleID := matches[1]
	reason := ""
	if len(matches) >= 3 {
		reason = strings.TrimSpace(matches[2])
	}

	return Suppression{
		RuleID: ruleID,
		Reason: reason,
		File:   file,
		Line:   line,
		Valid:  reason != "",
	}, true
}

// FilterViolations removes violations matched by valid suppressions.
// Invalid suppressions (empty reason) or attempts to suppress non-suppressibles
// return explicit suppression error violations.
func FilterViolations(violations []Violation, suppressions []Suppression) []Violation {
	var remaining []Violation

	// Index valid suppressions by File:Line:RuleID
	type key struct {
		file   string
		line   int
		ruleID string
	}

	exactSuppressed := make(map[key]bool)

	for _, s := range suppressions {
		if NonSuppressibleRules[s.RuleID] {
			remaining = append(remaining, Violation{
				RuleID:  s.RuleID,
				Code:    "LOY-ARCH-099",
				Message: "non-suppressible rule violation",
				Detail:  "rule " + s.RuleID + " cannot be suppressed with loy:ignore",
				Hint:    "refactor code to satisfy the architectural invariant",
				File:    s.File,
				Line:    s.Line,
			})
			continue
		}

		if !s.Valid {
			remaining = append(remaining, Violation{
				RuleID:  s.RuleID,
				Code:    "LOY-ARCH-099",
				Message: "invalid suppression directive",
				Detail:  "loy:ignore " + s.RuleID + " requires a non-empty reason=\"...\"",
				Hint:    "add reason=\"explanation\" to the ignore comment",
				File:    s.File,
				Line:    s.Line,
			})
			continue
		}

		exactSuppressed[key{file: s.File, line: s.Line, ruleID: s.RuleID}] = true
		exactSuppressed[key{file: s.File, line: s.Line + 1, ruleID: s.RuleID}] = true // comment directly above statement
	}

	for _, v := range violations {
		if exactSuppressed[key{file: v.File, line: v.Line, ruleID: v.RuleID}] {
			continue
		}
		remaining = append(remaining, v)
	}

	return remaining
}
