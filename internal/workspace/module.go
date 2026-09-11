package workspace

// ModuleType represents the classification of a Go module.
type ModuleType string

const (
	TypeApp     ModuleType = "app"
	TypePackage ModuleType = "package"
)

// Module represents an individual Go module in a project or workspace.
type Module struct {
	Name               string     `json:"name"`
	Path               string     `json:"path"`
	Rel                string     `json:"rel"`
	Type               ModuleType `json:"type"`
	DirectDependencies []string   `json:"direct_dependencies"`
}
