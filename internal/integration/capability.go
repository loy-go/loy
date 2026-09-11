package integration

// Capability represents an infrastructure or platform capability category.
type Capability string

const (
	CapHTTP      Capability = "http"
	CapDatabase  Capability = "database"
	CapCache     Capability = "cache"
	CapQueue     Capability = "queue"
	CapRPC       Capability = "rpc"
	CapTemplate  Capability = "template"
	CapTelemetry Capability = "telemetry"
	CapAssets    Capability = "assets"
)

// String returns the capability string identifier.
func (c Capability) String() string {
	return string(c)
}

// IsValid checks whether the capability is a known Loy capability.
func (c Capability) IsValid() bool {
	switch c {
	case CapHTTP, CapDatabase, CapCache, CapQueue, CapRPC, CapTemplate, CapTelemetry, CapAssets:
		return true
	default:
		return false
	}
}
