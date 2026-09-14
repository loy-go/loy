package plan

import (
	"bytes"
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/loy-go/loy/internal/diagnostics"
	"github.com/loy-go/loy/internal/filesystem"
	"github.com/loy-go/loy/internal/generator"
	"github.com/loy-go/loy/internal/generator/model"
)

// ConflictDetector inspects proposed artifacts against the current filesystem state.
type ConflictDetector struct {
	fs filesystem.FileSystem
}

// NewConflictDetector constructs a ConflictDetector.
func NewConflictDetector(fs filesystem.FileSystem) *ConflictDetector {
	return &ConflictDetector{fs: fs}
}

// ResolveTargetPath returns the relative path from target root, validating path jail.
func ResolveTargetPath(targetDir, relPath string) (string, error) {
	cleanPath := filepath.Clean(relPath)
	if strings.HasPrefix(cleanPath, "/") {
		return "", fmt.Errorf("absolute paths not allowed: %s", relPath)
	}
	if cleanPath == ".." || strings.HasPrefix(cleanPath, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("path escapes target root: %s", relPath)
	}

	if targetDir == "" || targetDir == "." {
		return cleanPath, nil
	}
	return filepath.Clean(filepath.Join(targetDir, cleanPath)), nil
}

// Check evaluates an artifact against filesystem state and determines the safe OperationType.
// Returns an error/diagnostic if an unresolvable conflict occurs.
func (c *ConflictDetector) Check(ctx context.Context, targetDir string, artifact model.Artifact, opts generator.Options) (OperationType, error) {
	relPath, err := ResolveTargetPath(targetDir, artifact.Path)
	if err != nil {
		diag := diagnostics.NewError(
			diagnostics.CodeFSPathTraversal,
			fmt.Sprintf("path validation failed for %q: %v", artifact.Path, err),
		)
		diag.File = artifact.Path
		return "", &diag
	}

	exists, err := c.fs.Exists(relPath)
	if err != nil {
		return "", fmt.Errorf("checking path existence %s: %w", relPath, err)
	}

	// 1. File does not exist -> OpCreate
	if !exists {
		return OpCreate, nil
	}

	// Read existing content
	existingData, err := c.fs.ReadFile(relPath)
	if err != nil {
		return "", fmt.Errorf("reading existing file %s: %w", relPath, err)
	}

	// 2. MixedOwned file with managed region -> OpSplice
	if artifact.Ownership == model.MixedOwned && artifact.Region != "" {
		return OpSplice, nil
	}

	// 3. Exact identical content -> OpSkip
	if bytes.Equal(existingData, artifact.Content) {
		return OpSkip, nil
	}

	// 4. DeveloperOwned file collision
	if artifact.Ownership == model.DeveloperOwned {
		if opts.Force {
			return OpOverwrite, nil
		}

		diag := diagnostics.NewError(
			diagnostics.CodeGenConflict,
			fmt.Sprintf("cannot overwrite developer-owned file %q without --force", artifact.Path),
		)
		diag.File = artifact.Path
		diag.Hint = "Provide --force to overwrite existing developer-owned file, or resolve manually"
		return "", &diag
	}

	// 5. GeneratedOwned file collision
	if artifact.Ownership == model.GeneratedOwned {
		// If --force is supplied, always overwrite
		if opts.Force {
			return OpOverwrite, nil
		}

		// Verify header marker to ensure developer did not modify this generated file
		hasHeader := strings.Contains(string(existingData), model.GeneratedFileHeader)
		if !hasHeader {
			diag := diagnostics.NewError(
				diagnostics.CodeGenConflict,
				fmt.Sprintf("generated file %q was modified by developer (missing standard header)", artifact.Path),
			)
			diag.File = artifact.Path
			diag.Hint = "Run with --force to overwrite changes, or restore standard header"
			return "", &diag
		}

		return OpOverwrite, nil
	}

	// Default fallback to overwrite if force, else conflict
	if opts.Force {
		return OpOverwrite, nil
	}

	diag := diagnostics.NewError(
		diagnostics.CodeGenConflict,
		fmt.Sprintf("conflicting existing file %q", artifact.Path),
	)
	diag.File = artifact.Path
	return "", &diag
}
