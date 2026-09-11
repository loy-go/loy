package discovery

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/uloydev/loy/internal/diagnostics"
	"github.com/uloydev/loy/internal/filesystem"
	"github.com/uloydev/loy/internal/process"
	"golang.org/x/mod/modfile"
)

// Discoverer identifies the active Loy project or workspace root.
type Discoverer struct {
	fs     filesystem.FileSystem
	runner process.Runner
}

// NewDiscoverer constructs a new Discoverer instance with explicit dependencies.
func NewDiscoverer(fs filesystem.FileSystem, runner process.Runner) (*Discoverer, error) {
	if fs == nil {
		return nil, fmt.Errorf("filesystem cannot be nil")
	}
	return &Discoverer{
		fs:     fs,
		runner: runner,
	}, nil
}

// Discover traverses up from startDir to find the project or workspace root.
// Precedence per Doc 11:
// 1. Current directory / closest parent with loy.yaml (project root)
// 2. Closest parent with go.mod
// 3. Ancestor with go.work (workspace context)
func (d *Discoverer) Discover(ctx context.Context, startDir string) (*DiscoveredContext, *diagnostics.Diagnostic) {
	cleanStart := filepath.Clean(startDir)
	curr := cleanStart

	var foundManifestDir string
	var foundGoModDir string
	var foundGoWorkDir string

	for {
		manifestPath := filepath.Join(curr, "loy.yaml")
		if exists, _ := d.fs.Exists(manifestPath); exists && foundManifestDir == "" {
			foundManifestDir = curr
		}

		goModPath := filepath.Join(curr, "go.mod")
		if exists, _ := d.fs.Exists(goModPath); exists && foundGoModDir == "" {
			foundGoModDir = curr
		}

		goWorkPath := filepath.Join(curr, "go.work")
		if exists, _ := d.fs.Exists(goWorkPath); exists && foundGoWorkDir == "" {
			foundGoWorkDir = curr
		}

		// Check if we hit git boundary or root
		gitPath := filepath.Join(curr, ".git")
		hasGit, _ := d.fs.Exists(gitPath)

		parent := filepath.Dir(curr)
		if parent == curr || hasGit {
			break
		}
		curr = parent
	}

	// Fallback to go env if runner is available and not found yet
	if foundGoModDir == "" && d.runner != nil {
		if res, err := d.runner.Run(ctx, cleanStart, "go", "env", "GOMOD"); err == nil && res.Success() {
			modFile := strings.TrimSpace(string(res.Stdout))
			if modFile != "" && modFile != "/dev/null" && modFile != "NUL" {
				foundGoModDir = filepath.Dir(modFile)
			}
		}
	}

	if foundGoWorkDir == "" && d.runner != nil {
		if res, err := d.runner.Run(ctx, cleanStart, "go", "env", "GOWORK"); err == nil && res.Success() {
			workFile := strings.TrimSpace(string(res.Stdout))
			if workFile != "" && workFile != "/dev/null" && workFile != "NUL" {
				foundGoWorkDir = filepath.Dir(workFile)
			}
		}
	}

	var rootDir string
	if foundManifestDir != "" {
		rootDir = foundManifestDir
	} else if foundGoModDir != "" {
		rootDir = foundGoModDir
	} else if foundGoWorkDir != "" {
		rootDir = foundGoWorkDir
	} else {
		return nil, &diagnostics.Diagnostic{
			Severity: diagnostics.SeverityError,
			Code:     diagnostics.CodeProjectRootNotFound,
			Message:  "no Loy project or Go module found in directory or its parents",
			Hint:     "run 'loy init' to initialize a project or 'loy new <name>' to create one",
			File:     cleanStart,
		}
	}

	res := &DiscoveredContext{
		RootDir: rootDir,
	}

	manifestPath := filepath.Join(rootDir, "loy.yaml")
	if exists, _ := d.fs.Exists(manifestPath); exists {
		res.HasManifest = true
		res.ManifestPath = manifestPath
	}

	goModPath := filepath.Join(rootDir, "go.mod")
	if exists, _ := d.fs.Exists(goModPath); exists {
		res.HasGoMod = true
		res.GoModPath = goModPath
	}

	// Check if this project is part of an outer or local workspace
	effectiveWorkDir := foundGoWorkDir
	if effectiveWorkDir == "" {
		effectiveWorkDir = rootDir
	}
	goWorkPath := filepath.Join(effectiveWorkDir, "go.work")
	if exists, _ := d.fs.Exists(goWorkPath); exists {
		res.HasGoWork = true
		res.GoWorkPath = goWorkPath
		res.IsWorkspace = true
	}

	return res, nil
}

// InferProjectName reads module path from go.mod in targetDir and returns the base package name.
func (d *Discoverer) InferProjectName(targetDir string) (string, error) {
	goModPath := filepath.Join(targetDir, "go.mod")
	data, err := d.fs.ReadFile(goModPath)
	if err != nil {
		return "", err
	}
	f, err := modfile.Parse(goModPath, data, nil)
	if err != nil || f.Module == nil {
		return "", fmt.Errorf("parsing go.mod: %w", err)
	}
	parts := strings.Split(strings.TrimRight(f.Module.Mod.Path, "/"), "/")
	return parts[len(parts)-1], nil
}
