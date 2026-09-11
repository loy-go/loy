package builtin

import (
	"strings"
)

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
