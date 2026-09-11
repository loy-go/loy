package filesystem

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
)

var (
	// ErrPathTraversal is returned when a path resolves outside the root boundary.
	ErrPathTraversal = errors.New("path traversal detected: target path escapes root directory")
)

// CleanAndValidatePath verifies that target is safely within rootDir and returns the cleaned absolute path.
func CleanAndValidatePath(rootDir, targetPath string) (string, error) {
	if rootDir == "" {
		return "", errors.New("root directory cannot be empty")
	}

	absRoot, err := filepath.Abs(rootDir)
	if err != nil {
		return "", fmt.Errorf("resolving root path: %w", err)
	}

	var candidate string
	if filepath.IsAbs(targetPath) {
		candidate = filepath.Clean(targetPath)
	} else {
		candidate = filepath.Clean(filepath.Join(absRoot, targetPath))
	}

	rel, err := filepath.Rel(absRoot, candidate)
	if err != nil {
		return "", fmt.Errorf("computing relative path: %w", err)
	}

	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("%w: %s escapes %s", ErrPathTraversal, targetPath, rootDir)
	}

	return candidate, nil
}
