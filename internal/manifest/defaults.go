package manifest

// NormalizedConfig is the validated and normalized manifest ready for consumption.
type NormalizedConfig struct {
	Version      int
	Project      ProjectConfig
	Defaults     DefaultsConfig
	Integrations map[string]Integration
	Architecture ArchitectureConfig
	Workspace    WorkspaceConfig
}

// Normalize takes a validated Manifest and applies built-in default fallbacks.
func Normalize(m *Manifest) *NormalizedConfig {
	norm := &NormalizedConfig{
		Version:      m.Version,
		Project:      m.Project,
		Defaults:     m.Defaults,
		Integrations: make(map[string]Integration),
		Architecture: m.Architecture,
		Workspace:    m.Workspace,
	}

	for k, v := range m.Integrations {
		norm.Integrations[k] = v
	}

	// Apply canonical defaults if completely unset
	if norm.Defaults.HTTP == "" {
		norm.Defaults.HTTP = "fiber"
	}
	if norm.Defaults.Database == "" {
		norm.Defaults.Database = "postgres"
	}
	if norm.Defaults.Cache == "" {
		norm.Defaults.Cache = "valkey"
	}
	if norm.Defaults.Queue == "" {
		norm.Defaults.Queue = "asynq"
	}

	return norm
}
