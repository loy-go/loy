package workspace

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/uloydev/loy/internal/diagnostics"
	"github.com/uloydev/loy/internal/filesystem"
	"golang.org/x/mod/modfile"
)

// Resolver reads and resolves Go workspace topology.
type Resolver struct {
	fs filesystem.FileSystem
}

// NewResolver creates a new workspace Resolver.
func NewResolver(fs filesystem.FileSystem) (*Resolver, error) {
	if fs == nil {
		return nil, fmt.Errorf("filesystem cannot be nil")
	}
	return &Resolver{fs: fs}, nil
}

// Resolve inspects a rootDir containing go.work or a single go.mod and models the workspace.
func (r *Resolver) Resolve(rootDir string) (*Workspace, []*diagnostics.Diagnostic) {
	ws := &Workspace{
		RootDir: rootDir,
		Modules: make(map[string]*Module),
	}

	cleanRoot, err := filesystem.CleanAndValidatePath(rootDir, rootDir)
	if err != nil {
		return nil, []*diagnostics.Diagnostic{{
			Severity: diagnostics.SeverityError,
			Code:     diagnostics.CodeFSPathTraversal,
			Message:  fmt.Sprintf("validating workspace root: %v", err),
			File:     rootDir,
		}}
	}

	goWorkPath, err := filesystem.CleanAndValidatePath(cleanRoot, filepath.Join(cleanRoot, "go.work"))
	hasWork := false
	if err == nil {
		hasWork, _ = r.fs.Exists(goWorkPath)
	}

	if hasWork {
		data, err := r.fs.ReadFile(goWorkPath)
		if err != nil {
			return nil, []*diagnostics.Diagnostic{{
				Severity: diagnostics.SeverityError,
				Code:     diagnostics.CodeWorkspaceConflict,
				Message:  fmt.Sprintf("reading go.work: %v", err),
				File:     goWorkPath,
			}}
		}

		workFile, err := modfile.ParseWork(goWorkPath, data, nil)
		if err != nil {
			return nil, []*diagnostics.Diagnostic{{
				Severity: diagnostics.SeverityError,
				Code:     diagnostics.CodeWorkspaceConflict,
				Message:  fmt.Sprintf("parsing go.work: %v", err),
				File:     goWorkPath,
			}}
		}

		for _, use := range workFile.Use {
			relPath := filepath.Clean(use.Path)
			absModDir, err := filesystem.CleanAndValidatePath(cleanRoot, filepath.Join(cleanRoot, relPath))
			if err != nil {
				return nil, []*diagnostics.Diagnostic{{
					Severity: diagnostics.SeverityError,
					Code:     diagnostics.CodeFSPathTraversal,
					Message:  fmt.Sprintf("module %s escapes workspace root: %v", relPath, err),
					File:     goWorkPath,
				}}
			}

			mod, diag := r.inspectModule(absModDir, relPath)
			if diag != nil {
				return nil, []*diagnostics.Diagnostic{diag}
			}
			if _, exists := ws.Modules[mod.Name]; exists {
				return nil, []*diagnostics.Diagnostic{{
					Severity: diagnostics.SeverityError,
					Code:     diagnostics.CodeWorkspaceConflict,
					Message:  fmt.Sprintf("conflicting module name %q declared at %s", mod.Name, mod.Rel),
					File:     goWorkPath,
				}}
			}
			ws.Modules[mod.Name] = mod
			if mod.Type == TypeApp {
				ws.Apps = append(ws.Apps, mod)
			} else {
				ws.Packages = append(ws.Packages, mod)
			}
		}
		return ws, nil
	}

	// Single module fallback (go.mod)
	goModPath, err := filesystem.CleanAndValidatePath(cleanRoot, filepath.Join(cleanRoot, "go.mod"))
	hasMod := false
	if err == nil {
		hasMod, _ = r.fs.Exists(goModPath)
	}

	if hasMod {
		mod, diag := r.inspectModule(cleanRoot, ".")
		if diag != nil {
			return nil, []*diagnostics.Diagnostic{diag}
		}
		ws.Modules[mod.Name] = mod
		if mod.Type == TypeApp {
			ws.Apps = append(ws.Apps, mod)
		} else {
			ws.Packages = append(ws.Packages, mod)
		}
		return ws, nil
	}

	return nil, []*diagnostics.Diagnostic{{
		Severity: diagnostics.SeverityError,
		Code:     diagnostics.CodeProjectRootNotFound,
		Message:  "neither go.work nor go.mod found in root directory",
		File:     cleanRoot,
	}}
}

func (r *Resolver) inspectModule(absModDir, relPath string) (*Module, *diagnostics.Diagnostic) {
	goModPath, err := filesystem.CleanAndValidatePath(absModDir, filepath.Join(absModDir, "go.mod"))
	if err != nil {
		return nil, &diagnostics.Diagnostic{
			Severity: diagnostics.SeverityError,
			Code:     diagnostics.CodeFSPathTraversal,
			Message:  fmt.Sprintf("validating module %s path: %v", relPath, err),
			File:     absModDir,
		}
	}

	data, err := r.fs.ReadFile(goModPath)
	if err != nil {
		return nil, &diagnostics.Diagnostic{
			Severity: diagnostics.SeverityError,
			Code:     diagnostics.CodeWorkspaceConflict,
			Message:  fmt.Sprintf("module %s is missing go.mod: %v", relPath, err),
			File:     goModPath,
		}
	}

	f, err := modfile.Parse(goModPath, data, nil)
	if err != nil || f.Module == nil {
		return nil, &diagnostics.Diagnostic{
			Severity: diagnostics.SeverityError,
			Code:     diagnostics.CodeWorkspaceConflict,
			Message:  fmt.Sprintf("parsing go.mod for module %s: %v", relPath, err),
			File:     goModPath,
		}
	}

	modName := f.Module.Mod.Path

	var directDeps []string
	for _, req := range f.Require {
		if req != nil && !req.Indirect {
			directDeps = append(directDeps, req.Mod.Path)
		}
	}

	modType := r.classifyModule(absModDir, relPath)

	return &Module{
		Name:               modName,
		Path:               absModDir,
		Rel:                relPath,
		Type:               modType,
		DirectDependencies: directDeps,
	}, nil
}

func (r *Resolver) classifyModule(absModDir, relPath string) ModuleType {
	if strings.HasPrefix(relPath, "apps/") || strings.Contains(relPath, "/apps/") {
		return TypeApp
	}
	cmdPath, err := filesystem.CleanAndValidatePath(absModDir, filepath.Join(absModDir, "cmd"))
	if err == nil {
		if exists, _ := r.fs.Exists(cmdPath); exists {
			return TypeApp
		}
	}
	mainPath, err := filesystem.CleanAndValidatePath(absModDir, filepath.Join(absModDir, "main.go"))
	if err == nil {
		if exists, _ := r.fs.Exists(mainPath); exists {
			return TypeApp
		}
	}
	return TypePackage
}
