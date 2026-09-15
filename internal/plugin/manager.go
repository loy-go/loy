package plugin

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	iofs "io/fs"
	"path/filepath"
	"strings"

	"github.com/loy-go/loy/internal/filesystem"
	"github.com/loy-go/loy/internal/generator"
	"github.com/loy-go/loy/internal/generator/model"
	"github.com/loy-go/loy/internal/generator/plan"
	"github.com/loy-go/loy/internal/process"
)

// Manager coordinates plugin discovery, execution, and artifact sandboxing.
type Manager struct {
	fs     filesystem.FileSystem
	runner process.Runner
	root   string // Project root directory
}

// NewManager constructs a Plugin Manager for the specified project root.
func NewManager(fs filesystem.FileSystem, runner process.Runner, projectRoot string) *Manager {
	if projectRoot == "" {
		projectRoot = "."
	}
	absRoot, err := filepath.Abs(projectRoot)
	if err == nil {
		projectRoot = absRoot
	}
	return &Manager{
		fs:     fs,
		runner: runner,
		root:   projectRoot,
	}
}

// PluginsDir returns the path to the project-local plugin storage directory (.loy/plugins).
func (m *Manager) PluginsDir() string {
	return filepath.Join(m.root, ".loy", "plugins")
}

// List scans .loy/plugins/ and returns all discovered and valid plugin manifests.
func (m *Manager) List(ctx context.Context) ([]*PluginManifest, error) {
	dir := m.PluginsDir()
	exists, err := m.fs.Exists(dir)
	if err != nil || !exists {
		return []*PluginManifest{}, nil
	}

	var manifests []*PluginManifest
	_ = m.fs.Walk(dir, func(path string, d iofs.DirEntry, err error) error {
		if err != nil || d == nil {
			return nil
		}
		if d.IsDir() {
			rel, err := filepath.Rel(dir, path)
			if err == nil && rel != "." && strings.Count(rel, string(filepath.Separator)) > 0 {
				return iofs.SkipDir
			}
			return nil
		}
		if d.Name() == "loy-plugin.yaml" || d.Name() == "loy-plugin.yml" {
			data, err := m.fs.ReadFile(path)
			if err == nil {
				if pm, err := ParsePluginManifest(data); err == nil && pm != nil {
					manifests = append(manifests, pm)
				}
			}
		}
		return nil
	})

	return manifests, nil
}

// Get loads a specific plugin manifest by name from .loy/plugins/<name>.
func (m *Manager) Get(name string) (*PluginManifest, string, error) {
	if !pluginNameRegex.MatchString(name) || filepath.Base(name) != name {
		return nil, "", fmt.Errorf("invalid plugin name: %q", name)
	}

	pluginDir, err := filesystem.CleanAndValidatePath(m.PluginsDir(), filepath.Join(m.PluginsDir(), name))
	if err != nil {
		return nil, "", fmt.Errorf("invalid plugin path: %w", err)
	}

	exists, err := m.fs.Exists(pluginDir)
	if err != nil || !exists {
		return nil, "", fmt.Errorf("plugin %q is not installed in %s", name, m.PluginsDir())
	}

	manifestPath := filepath.Join(pluginDir, "loy-plugin.yaml")
	data, err := m.fs.ReadFile(manifestPath)
	if err != nil {
		manifestPath = filepath.Join(pluginDir, "loy-plugin.yml")
		data, err = m.fs.ReadFile(manifestPath)
		if err != nil {
			return nil, "", fmt.Errorf("plugin %q missing loy-plugin.yaml: %w", name, err)
		}
	}

	pm, err := ParsePluginManifest(data)
	if err != nil {
		return nil, "", fmt.Errorf("invalid plugin %q manifest: %w", name, err)
	}

	return pm, pluginDir, nil
}

// Execute runs a plugin command in an isolated subprocess and applies generated artifacts safely.
func (m *Manager) Execute(ctx context.Context, pluginName, cmdName string, args []string, moduleName string, defaults map[string]string, dryRun, force bool) ([]model.Artifact, error) {
	pm, pluginDir, err := m.Get(pluginName)
	if err != nil {
		return nil, err
	}

	var cmdSpec *PluginCommandSpec
	for _, c := range pm.Commands {
		if c.Name == cmdName {
			cmdSpec = &c
			break
		}
	}
	if cmdSpec == nil {
		return nil, fmt.Errorf("plugin %q does not provide command %q", pluginName, cmdName)
	}

	execPath := filepath.Join(pluginDir, cmdSpec.Executable)

	payload := InvocationPayload{
		ProjectRoot: m.root,
		ModuleName:  moduleName,
		Command:     cmdName,
		Args:        args,
		Defaults:    defaults,
	}

	inputBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshaling invocation payload: %w", err)
	}

	res, err := m.runner.RunWithInput(ctx, pluginDir, bytes.NewReader(inputBytes), execPath, args...)
	if err != nil {
		return nil, fmt.Errorf("executing plugin command %q: %w", cmdName, err)
	}
	if res.ExitCode != 0 {
		return nil, fmt.Errorf("plugin command failed (exit code %d): %s", res.ExitCode, strings.TrimSpace(string(res.Stderr)))
	}

	var resp InvocationResponse
	if err := json.Unmarshal(res.Stdout, &resp); err != nil {
		return nil, fmt.Errorf("invalid response from plugin (expected JSON): %w", err)
	}

	if resp.Status == "error" || resp.Error != "" {
		return nil, fmt.Errorf("plugin error: %s", resp.Error)
	}

	// Security Sandbox Validation:
	// Verify that all artifact paths are inside project root and convert them to model.Artifact.
	var artifacts []model.Artifact
	for _, dto := range resp.Artifacts {
		if dto.Path == "" {
			return nil, fmt.Errorf("plugin emitted artifact with empty path")
		}

		// Enforce path sandbox
		cleanPath, err := filesystem.CleanAndValidatePath(m.root, filepath.Join(m.root, dto.Path))
		if err != nil {
			return nil, fmt.Errorf("security violation: plugin emitted path outside project root %q: %w", dto.Path, err)
		}

		relPath, err := filepath.Rel(m.root, cleanPath)
		if err != nil || relPath == "" || strings.HasPrefix(relPath, "..") {
			return nil, fmt.Errorf("security violation: path traversal detected: %s", dto.Path)
		}

		cleanRel := filepath.Clean(relPath)
		if cleanRel == ".git" || strings.HasPrefix(cleanRel, ".git"+string(filepath.Separator)) {
			return nil, fmt.Errorf("security violation: plugin cannot write to .git directory")
		}

		ownership := model.DeveloperOwned
		switch strings.ToLower(dto.Ownership) {
		case "generated":
			ownership = model.GeneratedOwned
		case "mixed":
			ownership = model.MixedOwned
		}

		perm := dto.Permissions
		if perm == 0 {
			perm = 0644
		}

		artifacts = append(artifacts, model.Artifact{
			Path:        relPath,
			Content:     []byte(dto.Content),
			Ownership:   ownership,
			Region:      dto.Region,
			Permissions: iofs.FileMode(perm),
		})
	}

	if dryRun {
		return artifacts, nil
	}

	// Apply artifacts using atomic plan engine
	builder := plan.NewBuilder(m.fs)
	p, err := builder.Build(ctx, m.root, artifacts, generator.Options{Force: force})
	if err != nil {
		return nil, fmt.Errorf("building execution plan: %w", err)
	}

	executor := plan.NewExecutor(m.fs)
	if err := executor.Execute(ctx, p); err != nil {
		return nil, fmt.Errorf("applying plugin artifacts: %w", err)
	}

	return artifacts, nil
}

// Install copies a local plugin directory or clones a git repo into .loy/plugins/<name>.
func (m *Manager) Install(ctx context.Context, source string) (*PluginManifest, error) {
	if strings.HasPrefix(source, "-") {
		return nil, fmt.Errorf("invalid plugin source: options not allowed")
	}

	pluginsDir := m.PluginsDir()
	if err := m.fs.MkdirAll(pluginsDir, 0755); err != nil {
		return nil, fmt.Errorf("creating plugins directory: %w", err)
	}

	// Check if source is a local directory containing loy-plugin.yaml
	manifestPath := filepath.Join(source, "loy-plugin.yaml")
	data, err := m.fs.ReadFile(manifestPath)
	if err != nil {
		manifestPath = filepath.Join(source, "loy-plugin.yml")
		data, err = m.fs.ReadFile(manifestPath)
	}

	if err == nil {
		pm, parseErr := ParsePluginManifest(data)
		if parseErr != nil {
			return nil, fmt.Errorf("invalid local plugin manifest: %w", parseErr)
		}

		destDir := filepath.Join(pluginsDir, pm.Name)
		_ = m.fs.RemoveAll(destDir)
		if err := m.fs.MkdirAll(destDir, 0755); err != nil {
			return nil, fmt.Errorf("creating plugin destination: %w", err)
		}

		// Recursively copy source directory into destDir
		_ = m.fs.Walk(source, func(path string, d iofs.DirEntry, walkErr error) error {
			if walkErr != nil || d == nil {
				return nil
			}
			rel, relErr := filepath.Rel(source, path)
			if relErr != nil || rel == "." {
				return nil
			}
			target := filepath.Join(destDir, rel)
			if d.IsDir() {
				return m.fs.MkdirAll(target, 0755)
			}
			fBytes, readErr := m.fs.ReadFile(path)
			if readErr != nil {
				return readErr
			}
			info, statErr := d.Info()
			perm := iofs.FileMode(0644)
			if statErr == nil {
				perm = info.Mode()
			}
			return m.fs.WriteFile(target, fBytes, perm)
		})

		return pm, nil
	}

	// Git clone installation
	// Target folder name from URL
	parts := strings.Split(strings.TrimSuffix(source, ".git"), "/")
	repoName := parts[len(parts)-1]
	if repoName == "" {
		repoName = "plugin"
	}

	tmpCloneDir := filepath.Join(pluginsDir, ".tmp-"+repoName)
	defer func() { _ = m.fs.RemoveAll(tmpCloneDir) }()

	res, err := m.runner.Run(ctx, pluginsDir, "git", "clone", "--depth=1", source, tmpCloneDir)
	if err != nil || res.ExitCode != 0 {
		return nil, fmt.Errorf("git clone failed for %s: %s", source, strings.TrimSpace(string(res.Stderr)))
	}

	mPath := filepath.Join(tmpCloneDir, "loy-plugin.yaml")
	mData, err := m.fs.ReadFile(mPath)
	if err != nil {
		mPath = filepath.Join(tmpCloneDir, "loy-plugin.yml")
		mData, err = m.fs.ReadFile(mPath)
		if err != nil {
			return nil, fmt.Errorf("cloned repository does not contain loy-plugin.yaml: %w", err)
		}
	}

	pm, err := ParsePluginManifest(mData)
	if err != nil {
		return nil, fmt.Errorf("invalid cloned plugin manifest: %w", err)
	}

	finalDest := filepath.Join(pluginsDir, pm.Name)
	_ = m.fs.RemoveAll(finalDest)
	if err := m.fs.MkdirAll(finalDest, 0755); err != nil {
		return nil, fmt.Errorf("creating final plugin directory: %w", err)
	}

	// Copy all files from tmpCloneDir to finalDest
	_ = m.fs.Walk(tmpCloneDir, func(path string, d iofs.DirEntry, walkErr error) error {
		if walkErr != nil || d == nil {
			return nil
		}
		rel, relErr := filepath.Rel(tmpCloneDir, path)
		if relErr != nil || rel == "." {
			return nil
		}
		target := filepath.Join(finalDest, rel)
		if d.IsDir() {
			return m.fs.MkdirAll(target, 0755)
		}
		fBytes, readErr := m.fs.ReadFile(path)
		if readErr != nil {
			return readErr
		}
		info, statErr := d.Info()
		perm := iofs.FileMode(0644)
		if statErr == nil {
			perm = info.Mode()
		}
		return m.fs.WriteFile(target, fBytes, perm)
	})

	return pm, nil
}

