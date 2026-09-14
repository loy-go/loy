package architecture

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
