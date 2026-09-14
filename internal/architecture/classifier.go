package architecture

import (
	"path/filepath"
	"strings"
)

// Classifier maps import paths and file paths to Architectural layers.
type Classifier struct {
	moduleName string
	overrides  map[string]Layer
}

// NewClassifier creates a layer classifier for the target module.
func NewClassifier(moduleName string, overrides map[string]string) *Classifier {
	c := &Classifier{
		moduleName: moduleName,
		overrides:  make(map[string]Layer),
	}
	for pkg, lyr := range overrides {
		c.overrides[pkg] = Layer(lyr)
	}
	return c
}

// Classify determines the layer for a given package import path.
func (c *Classifier) Classify(importPath string) Layer {
	// Check overrides first
	if lyr, ok := c.overrides[importPath]; ok {
		return lyr
	}

	// Only classify packages within this module or relative internal paths
	relPath := importPath
	if c.moduleName != "" && strings.HasPrefix(importPath, c.moduleName) {
		relPath = strings.TrimPrefix(importPath, c.moduleName)
		relPath = strings.TrimPrefix(relPath, "/")
	}

	parts := strings.Split(filepath.ToSlash(relPath), "/")
	for i, part := range parts {
		switch part {
		case "transport", "handler", "handlers", "http", "grpc", "controller", "controllers", "views", "view":
			return LayerTransport
		case "application", "service", "services", "usecase", "usecases":
			return LayerApplication
		case "domain", "model", "models", "entity", "entities":
			return LayerDomain
		case "infrastructure", "infra", "platform", "repository", "repositories", "database", "store", "storage":
			return LayerInfrastructure
		}
		// Also support internal/layer pattern
		if part == "internal" && i+1 < len(parts) {
			switch parts[i+1] {
			case "transport", "handler", "handlers", "http", "grpc", "controller", "controllers", "views", "view":
				return LayerTransport
			case "application", "service", "services", "usecase", "usecases":
				return LayerApplication
			case "domain", "model", "models", "entity", "entities":
				return LayerDomain
			case "infrastructure", "infra", "platform", "repository", "repositories", "database", "store", "storage":
				return LayerInfrastructure
			}
		}
	}

	return LayerUnknown
}
