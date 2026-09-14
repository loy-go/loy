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
	if c.moduleName != "" {
		if strings.HasPrefix(importPath, c.moduleName) {
			relPath = strings.TrimPrefix(importPath, c.moduleName)
			relPath = strings.TrimPrefix(relPath, "/")
		} else if !strings.HasPrefix(importPath, "internal/") &&
			!strings.HasPrefix(importPath, "pkg/") &&
			!strings.HasPrefix(importPath, "cmd/") &&
			!strings.HasPrefix(importPath, "views/") {
			return LayerUnknown
		}
	}

	parts := strings.Split(filepath.ToSlash(relPath), "/")
	for i, part := range parts {
		switch part {
		case "cmd", "transport", "handler", "handlers", "http", "grpc", "controller", "controllers", "views", "view", "api", "apis", "rpc", "endpoint", "endpoints", "ws", "websocket":
			return LayerTransport
		case "application", "service", "services", "usecase", "usecases", "worker", "workers", "job", "jobs", "consumer", "consumers", "cron":
			return LayerApplication
		case "domain", "model", "models", "entity", "entities", "event", "events", "policy", "policies":
			return LayerDomain
		case "infrastructure", "infra", "repository", "repositories", "database", "store", "storage", "postgres", "sqlite", "valkey", "redis", "asynq", "river":
			return LayerInfrastructure
		case "platform", "pkg":
			return LayerPlatform
		}
		// Also support internal/layer pattern
		if part == "internal" && i+1 < len(parts) {
			switch parts[i+1] {
			case "cmd", "transport", "handler", "handlers", "http", "grpc", "controller", "controllers", "views", "view", "api", "apis", "rpc", "endpoint", "endpoints", "ws", "websocket":
				return LayerTransport
			case "application", "service", "services", "usecase", "usecases", "worker", "workers", "job", "jobs", "consumer", "consumers", "cron":
				return LayerApplication
			case "domain", "model", "models", "entity", "entities", "event", "events", "policy", "policies":
				return LayerDomain
			case "infrastructure", "infra", "repository", "repositories", "database", "store", "storage", "postgres", "sqlite", "valkey", "redis", "asynq", "river":
				return LayerInfrastructure
			case "platform", "pkg":
				return LayerPlatform
			}
		}
	}

	return LayerUnknown
}
