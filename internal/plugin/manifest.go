package plugin

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

var pluginNameRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]{1,64}$`)

// PluginManifest models loy-plugin.yaml.
type PluginManifest struct {
	Name        string              `yaml:"name" json:"name"`
	Version     string              `yaml:"version" json:"version"`
	Description string              `yaml:"description,omitempty" json:"description,omitempty"`
	Author      string              `yaml:"author,omitempty" json:"author,omitempty"`
	Commands    []PluginCommandSpec `yaml:"commands" json:"commands"`
}

// PluginCommandSpec describes an executable generator command exposed by the plugin.
type PluginCommandSpec struct {
	Name        string `yaml:"name" json:"name"`
	Description string `yaml:"description" json:"description"`
	Executable  string `yaml:"executable" json:"executable"`
}

// InvocationPayload is passed to the plugin subprocess via stdin as JSON.
type InvocationPayload struct {
	ProjectRoot string            `json:"project_root"`
	ModuleName  string            `json:"module_name"`
	Command     string            `json:"command"`
	Args        []string          `json:"args"`
	Defaults    map[string]string `json:"defaults"`
}

// InvocationResponse is read from the plugin subprocess stdout as JSON.
type InvocationResponse struct {
	Status    string              `json:"status"` // "ok" | "error"
	Error     string              `json:"error,omitempty"`
	Artifacts []PluginArtifactDTO `json:"artifacts,omitempty"`
}

// PluginArtifactDTO models an artifact emitted by an external plugin.
type PluginArtifactDTO struct {
	Path        string `json:"path"`
	Content     string `json:"content"`
	Ownership   string `json:"ownership"` // "developer", "generated", "mixed"
	Region      string `json:"region,omitempty"`
	Permissions int    `json:"permissions,omitempty"`
}

// ParsePluginManifest parses and strictly validates a loy-plugin.yaml payload.
func ParsePluginManifest(data []byte) (*PluginManifest, error) {
	var m PluginManifest
	if err := yaml.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("parsing plugin manifest: %w", err)
	}

	m.Name = strings.TrimSpace(m.Name)
	if m.Name == "" {
		return nil, fmt.Errorf("plugin manifest missing required 'name'")
	}
	if !pluginNameRegex.MatchString(m.Name) {
		return nil, fmt.Errorf("invalid plugin name %q: must be alphanumeric, hyphen, or underscore (1-64 chars)", m.Name)
	}

	if len(m.Commands) == 0 {
		return nil, fmt.Errorf("plugin %q must define at least one command in 'commands'", m.Name)
	}

	for i, cmd := range m.Commands {
		cmd.Name = strings.TrimSpace(cmd.Name)
		m.Commands[i].Name = cmd.Name
		if cmd.Name == "" {
			return nil, fmt.Errorf("command #%d missing 'name'", i+1)
		}
		if cmd.Executable == "" {
			return nil, fmt.Errorf("command %q missing 'executable'", cmd.Name)
		}
		cleanExec := filepath.Clean(cmd.Executable)
		if filepath.IsAbs(cmd.Executable) ||
			strings.HasPrefix(cleanExec, "..") ||
			strings.Contains(cleanExec, "/../") ||
			strings.Contains(cleanExec, `\..\`) ||
			strings.HasPrefix(cmd.Executable, "/") ||
			strings.HasPrefix(cmd.Executable, `\`) ||
			filepath.VolumeName(cmd.Executable) != "" {
			return nil, fmt.Errorf("command %q executable path %q must be relative and cannot contain traversal", cmd.Name, cmd.Executable)
		}
	}

	return &m, nil
}
