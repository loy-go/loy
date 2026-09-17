package manifest

// Manifest is the root schema representing loy.yaml (v1).
type Manifest struct {
	Version      int                    `yaml:"version"`
	Project      ProjectConfig          `yaml:"project"`
	Defaults     DefaultsConfig         `yaml:"defaults,omitempty"`
	MultiTenancy MultiTenancyConfig     `yaml:"multi_tenancy,omitempty"`
	Integrations map[string]Integration `yaml:"integrations,omitempty"`
	Architecture ArchitectureConfig     `yaml:"architecture,omitempty"`
	Workspace    WorkspaceConfig        `yaml:"workspace,omitempty"`
}

// MultiTenancyConfig configures tenant isolation strategies.
type MultiTenancyConfig struct {
	Enabled   bool   `yaml:"enabled"`
	Strategy  string `yaml:"strategy,omitempty"` // "rls" | "column"
	TenantKey string `yaml:"tenant_key,omitempty"`
}

// ProjectConfig contains metadata about the Loy project.
type ProjectConfig struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description,omitempty"`
	Module      string `yaml:"module,omitempty"`
}

// DefaultsConfig specifies default framework integrations.
type DefaultsConfig struct {
	HTTP     string `yaml:"http,omitempty"`
	Database string `yaml:"database,omitempty"`
	Cache    string `yaml:"cache,omitempty"`
	Queue    string `yaml:"queue,omitempty"`
	Template string `yaml:"template,omitempty"`
	Assets   string `yaml:"assets,omitempty"`
}

// Integration defines third-party or infrastructure settings.
type Integration struct {
	Driver  string                 `yaml:"driver,omitempty"`
	Enabled *bool                  `yaml:"enabled,omitempty"`
	Config  map[string]interface{} `yaml:"config,omitempty"`
}

// LayerConfig defines a topological layer, its matching path patterns, and what target layers it is allowed to import.
type LayerConfig struct {
	Allows []string `yaml:"allows,omitempty"`
	Match  []string `yaml:"match,omitempty"`
}

// ArchitectureConfig customizes rule enforcement.
type ArchitectureConfig struct {
	Strict   bool                   `yaml:"strict,omitempty"`
	Excluded []string               `yaml:"excluded,omitempty"`
	Pattern  string                 `yaml:"pattern,omitempty"` // "ddd" | "hexagonal" | "cqrs" | "custom"
	Layers   map[string]LayerConfig `yaml:"layers,omitempty"`
}

// WorkspaceConfig models workspace-level configuration.
type WorkspaceConfig struct {
	DefaultTarget string   `yaml:"default_target,omitempty"`
	Apps          []string `yaml:"apps,omitempty"`
	Packages      []string `yaml:"packages,omitempty"`
}
