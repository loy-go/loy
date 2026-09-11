package rules

import (
	"context"
	"fmt"
	"strings"

	"github.com/uloydev/loy/internal/architecture"
	"github.com/uloydev/loy/internal/diagnostics"
)

// RuleArch001 flags cyclic dependencies. Non-suppressible.
type RuleArch001 struct{}

func (r *RuleArch001) ID() string          { return "ARCH-001" }
func (r *RuleArch001) Description() string { return "Dependency cycles are forbidden" }

func (r *RuleArch001) Check(ctx context.Context, a *architecture.Analysis) []architecture.Violation {
	var violations []architecture.Violation
	cycles := a.Graph.FindCycles()

	for _, cycle := range cycles {
		chain := strings.Join(cycle.Path, " -> ")
		startPkg := cycle.Path[0]

		file := ""
		line := 1
		for _, e := range a.Graph.EdgesFrom(startPkg) {
			if len(cycle.Path) > 1 && e.To == cycle.Path[1] {
				file = e.File
				line = e.Line
				break
			}
		}

		violations = append(violations, architecture.Violation{
			RuleID:       r.ID(),
			Code:         diagnostics.CodeArchCycleDependency,
			Message:      fmt.Sprintf("cyclic dependency detected: %s", chain),
			Detail:       fmt.Sprintf("package %s forms an illegal dependency cycle", startPkg),
			Hint:         "break the cycle by extracting common types or using dependency inversion",
			File:         file,
			Line:         line,
			Suppressible: false,
		})
	}

	return violations
}
