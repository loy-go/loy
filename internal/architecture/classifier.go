package architecture

import (
	"path/filepath"
	"strings"
)

// PatternRule pairs a path glob/prefix pattern with an architectural layer.
type PatternRule struct {
	Pattern string
	Layer   Layer
}

// Classifier maps import paths and file paths to Architectural layers.
type Classifier struct {
	moduleName string
	overrides  map[string]Layer
	patterns   []PatternRule
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

// AddPattern appends a custom path pattern to layer mapping.
func (c *Classifier) AddPattern(pattern string, layer Layer) {
	c.patterns = append(c.patterns, PatternRule{Pattern: pattern, Layer: layer})
}

// SetPatterns sets all custom pattern rules.
func (c *Classifier) SetPatterns(patterns []PatternRule) {
	c.patterns = patterns
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

	// Check custom pattern rules next
	for _, p := range c.patterns {
		if matchPattern(p.Pattern, relPath) {
			return p.Layer
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

func matchPattern(pattern, path string) bool {
	pattern = filepath.ToSlash(pattern)
	path = filepath.ToSlash(path)

	if path == pattern {
		return true
	}
	if ok, _ := filepath.Match(pattern, path); ok {
		return true
	}

	patParts := strings.Split(pattern, "/")
	pathParts := strings.Split(path, "/")

	// If pattern ends with "**" (e.g. internal/*/service/**)
	if len(patParts) > 0 && patParts[len(patParts)-1] == "**" {
		prefixParts := patParts[:len(patParts)-1]
		if len(pathParts) >= len(prefixParts) {
			matched := true
			for i, p := range prefixParts {
				if ok, _ := filepath.Match(p, pathParts[i]); !ok {
					matched = false
					break
				}
			}
			if matched {
				return true
			}
		}
	}

	// If pattern ends with "*"
	if len(patParts) > 0 && patParts[len(patParts)-1] == "*" {
		prefixParts := patParts[:len(patParts)-1]
		if len(pathParts) == len(patParts) {
			matched := true
			for i, p := range prefixParts {
				if ok, _ := filepath.Match(p, pathParts[i]); !ok {
					matched = false
					break
				}
			}
			if matched {
				return true
			}
		}
	}

	if strings.HasPrefix(path, pattern+"/") {
		return true
	}

	return false
}

