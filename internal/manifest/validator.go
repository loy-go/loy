package manifest

import (
	"fmt"
	"regexp"

	"github.com/uloydev/loy/internal/diagnostics"
)

var (
	validNameRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

	validDrivers = map[string]map[string]bool{
		"http":     {"fiber": true, "chi": true, "echo": true, "nethttp": true},
		"database": {"postgres": true, "mysql": true, "sqlite": true, "none": true},
		"cache":    {"valkey": true, "redis": true, "memory": true, "none": true},
		"queue":    {"asynq": true, "river": true, "none": true},
		"template": {"templ": true, "html": true, "none": true},
		"assets":   {"vite": true, "none": true},
	}
)

// Validator validates semantic rules for a parsed Manifest.
type Validator struct{}

// NewValidator returns a new manifest Validator.
func NewValidator() *Validator {
	return &Validator{}
}

// Validate checks semantic validity of manifest fields.
func (v *Validator) Validate(filename string, m *Manifest) []*diagnostics.Diagnostic {
	var diags []*diagnostics.Diagnostic

	if m.Version != 1 {
		diags = append(diags, &diagnostics.Diagnostic{
			Severity: diagnostics.SeverityError,
			Code:     diagnostics.CodeConfigValidationError,
			Message:  fmt.Sprintf("unsupported manifest version %d (expected 1)", m.Version),
			Hint:     "set 'version: 1' in loy.yaml",
			File:     filename,
		})
	}

	if m.Project.Name == "" {
		diags = append(diags, &diagnostics.Diagnostic{
			Severity: diagnostics.SeverityError,
			Code:     diagnostics.CodeConfigValidationError,
			Message:  "project.name is required",
			Hint:     "specify a project name, e.g. 'project: { name: myapp }'",
			File:     filename,
		})
	} else if !validNameRegex.MatchString(m.Project.Name) {
		diags = append(diags, &diagnostics.Diagnostic{
			Severity: diagnostics.SeverityError,
			Code:     diagnostics.CodeConfigValidationError,
			Message:  fmt.Sprintf("invalid project.name %q: must contain only alphanumeric characters, dashes or underscores", m.Project.Name),
			Hint:     "use a valid project name like 'acme-api' or 'my_project'",
			File:     filename,
		})
	}

	// Validate default drivers if specified
	defaultsToCheck := map[string]string{
		"http":     m.Defaults.HTTP,
		"database": m.Defaults.Database,
		"cache":    m.Defaults.Cache,
		"queue":    m.Defaults.Queue,
		"template": m.Defaults.Template,
		"assets":   m.Defaults.Assets,
	}

	for capability, driver := range defaultsToCheck {
		if driver != "" && !validDrivers[capability][driver] {
			diags = append(diags, &diagnostics.Diagnostic{
				Severity: diagnostics.SeverityError,
				Code:     diagnostics.CodeConfigValidationError,
				Message:  fmt.Sprintf("unknown %s driver %q", capability, driver),
				Hint:     fmt.Sprintf("supported %s drivers: %v", capability, allowedDriversString(capability)),
				File:     filename,
			})
		}
	}

	// Validate integrations capabilities and drivers
	for name, integration := range m.Integrations {
		allowed, knownCap := validDrivers[name]
		if !knownCap {
			diags = append(diags, &diagnostics.Diagnostic{
				Severity: diagnostics.SeverityError,
				Code:     diagnostics.CodeConfigValidationError,
				Message:  fmt.Sprintf("unknown integration capability %q", name),
				Hint:     "supported capabilities: http, database, cache, queue, template, assets",
				File:     filename,
			})
			continue
		}
		if integration.Driver != "" && !allowed[integration.Driver] {
			diags = append(diags, &diagnostics.Diagnostic{
				Severity: diagnostics.SeverityError,
				Code:     diagnostics.CodeConfigValidationError,
				Message:  fmt.Sprintf("unknown driver %q for integration %q", integration.Driver, name),
				Hint:     fmt.Sprintf("supported drivers: %v", allowedDriversString(name)),
				File:     filename,
			})
		}
	}

	return diags
}

func allowedDriversString(capability string) string {
	drivers := validDrivers[capability]
	var list []string
	for d := range drivers {
		list = append(list, d)
	}
	return fmt.Sprintf("%v", list)
}
