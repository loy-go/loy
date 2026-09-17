package architecture

import "strings"

// Layer represents the architectural layer of a package.
type Layer string

const (
	LayerUnknown        Layer = "Unknown"
	LayerTransport      Layer = "Transport"
	LayerApplication    Layer = "Application"
	LayerDomain         Layer = "Domain"
	LayerInfrastructure Layer = "Infrastructure"
	LayerPlatform       Layer = "Platform"
)

// AllowedMatrix defines layer-to-layer import permissions per Spec 05.
// Transport -> Application, Domain, Transport, Platform
// Application -> Domain, Application, Platform
// Domain -> Domain, Platform
// Infrastructure -> Application, Domain, Infrastructure, Platform
// Platform -> Platform
var AllowedMatrix = map[Layer]map[Layer]bool{
	LayerTransport: {
		LayerTransport:   true,
		LayerApplication: true,
		LayerDomain:      true,
		LayerPlatform:    true,
	},
	LayerApplication: {
		LayerApplication: true,
		LayerDomain:      true,
		LayerPlatform:    true,
	},
	LayerDomain: {
		LayerDomain:   true,
		LayerPlatform: true,
	},
	LayerInfrastructure: {
		LayerInfrastructure: true,
		LayerApplication:    true,
		LayerDomain:         true,
		LayerPlatform:       true,
	},
	LayerPlatform: {
		LayerPlatform: true,
	},
}

// IsAllowedDirection returns true if fromLayer may import toLayer.
func IsAllowedDirection(from, to Layer) bool {
	if from == LayerUnknown || to == LayerUnknown || from == "" || to == "" {
		return true // skip unclassified packages, standard library, and external dependencies
	}
	if allowedTargets, ok := AllowedMatrix[from]; ok {
		return allowedTargets[to]
	}
	return true
}

// Topology represents a directed dependency topology between layers.
type Topology struct {
	custom        bool
	allowedMatrix map[Layer]map[Layer]bool
}

// NewDefaultTopology creates standard 4-layer DDD topology.
func NewDefaultTopology() *Topology {
	top := NewCustomTopology()
	for from, targets := range AllowedMatrix {
		for to, allowed := range targets {
			if allowed {
				top.Allow(from, to)
			}
		}
	}
	top.custom = false
	return top
}

// NewCustomTopology creates a user-defined layer DAG.
func NewCustomTopology() *Topology {
	return &Topology{
		custom:        true,
		allowedMatrix: make(map[Layer]map[Layer]bool),
	}
}

// IsCustom returns true if the topology was customized by user config.
func (t *Topology) IsCustom() bool {
	return t != nil && t.custom
}

// Allow registers permission for from layer to import to layer.
func (t *Topology) Allow(from, to Layer) {
	fromNorm := Layer(strings.ToLower(string(from)))
	toNorm := Layer(strings.ToLower(string(to)))
	if t.allowedMatrix == nil {
		t.allowedMatrix = make(map[Layer]map[Layer]bool)
	}
	if t.allowedMatrix[fromNorm] == nil {
		t.allowedMatrix[fromNorm] = make(map[Layer]bool)
	}
	t.allowedMatrix[fromNorm][toNorm] = true
}

// IsAllowed evaluates whether from may import to.
func (t *Topology) IsAllowed(from, to Layer) bool {
	if t == nil {
		return IsAllowedDirection(from, to)
	}
	if from == LayerUnknown || to == LayerUnknown || from == "" || to == "" {
		return true
	}
	fromNorm := Layer(strings.ToLower(string(from)))
	toNorm := Layer(strings.ToLower(string(to)))
	if fromNorm == toNorm {
		return true
	}
	if targets, ok := t.allowedMatrix[fromNorm]; ok {
		return targets[toNorm]
	}
	return false
}

// LayerTopologyConfig models layer topological permissions and match patterns.
type LayerTopologyConfig struct {
	Allows []string
	Match  []string
}

// BuildTopology constructs a Topology and PatternRule slice from pattern presets and custom layer definitions.
func BuildTopology(pattern string, layers map[string]LayerTopologyConfig) (*Topology, []PatternRule) {
	if len(layers) == 0 && pattern == "" {
		return nil, nil
	}

	topology := NewCustomTopology()
	var patterns []PatternRule

	switch strings.ToLower(pattern) {
	case "hexagonal":
		topology.Allow("adapter", "port")
		topology.Allow("adapter", "domain")
		topology.Allow("adapter", "platform")
		topology.Allow("port", "domain")
		topology.Allow("port", "platform")
		topology.Allow("domain", "platform")
		topology.Allow("platform", "platform")
	case "cqrs":
		topology.Allow("api", "command")
		topology.Allow("api", "query")
		topology.Allow("api", "metadata")
		topology.Allow("api", "platform")
		topology.Allow("command", "domain")
		topology.Allow("command", "database")
		topology.Allow("command", "event")
		topology.Allow("command", "platform")
		topology.Allow("query", "database")
		topology.Allow("query", "platform")
		topology.Allow("domain", "event")
		topology.Allow("domain", "platform")
		topology.Allow("database", "platform")
		topology.Allow("event", "platform")
		topology.Allow("platform", "platform")
	}

	for layerName, lCfg := range layers {
		lyr := Layer(layerName)
		for _, allowed := range lCfg.Allows {
			topology.Allow(lyr, Layer(allowed))
		}
		for _, match := range lCfg.Match {
			patterns = append(patterns, PatternRule{
				Pattern: match,
				Layer:   lyr,
			})
		}
	}

	return topology, patterns
}

