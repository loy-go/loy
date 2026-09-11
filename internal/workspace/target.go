package workspace

import (
	"fmt"
	"strings"

	"github.com/uloydev/loy/internal/diagnostics"
)

// SelectTarget matches a requested target against workspace apps.
func (ws *Workspace) SelectTarget(target string) (*Module, *diagnostics.Diagnostic) {
	if target == "" {
		if len(ws.Apps) == 1 {
			return ws.Apps[0], nil
		}
		if len(ws.Apps) > 1 {
			var appNames []string
			for _, a := range ws.Apps {
				appNames = append(appNames, a.Name)
			}
			return nil, &diagnostics.Diagnostic{
				Severity: diagnostics.SeverityError,
				Code:     diagnostics.CodeWorkspaceTargetError,
				Message:  fmt.Sprintf("multiple apps found in workspace; please specify --target [%s]", strings.Join(appNames, ", ")),
				Hint:     "pass --target=<app_name>",
				File:     ws.RootDir,
			}
		}
		// If only one module overall, return it
		if len(ws.Modules) == 1 {
			for _, m := range ws.Modules {
				return m, nil
			}
		}
		return nil, &diagnostics.Diagnostic{
			Severity: diagnostics.SeverityError,
			Code:     diagnostics.CodeWorkspaceTargetError,
			Message:  "no runnable app target found in workspace",
			File:     ws.RootDir,
		}
	}

	// Match by full name, rel path, or base name
	for _, app := range ws.Apps {
		if app.Name == target || app.Rel == target || strings.HasSuffix(app.Rel, "/"+target) {
			return app, nil
		}
	}

	// Check in all modules
	for _, mod := range ws.Modules {
		if mod.Name == target || mod.Rel == target || strings.HasSuffix(mod.Rel, "/"+target) {
			return mod, nil
		}
	}

	return nil, &diagnostics.Diagnostic{
		Severity: diagnostics.SeverityError,
		Code:     diagnostics.CodeWorkspaceTargetError,
		Message:  fmt.Sprintf("target %q not found in workspace", target),
		Hint:     "run 'loy graph' or inspect go.work to see active modules",
		File:     ws.RootDir,
	}
}
