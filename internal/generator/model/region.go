package model

// ManagedRegion represents an identifiable section inside a MixedOwned artifact.
type ManagedRegion struct {
	Name    string
	Content string
}

// RegionMarker constants for comment splicing.
const (
	RegionMarkerPrefix = "// loy:region:"
	RegionMarkerEnd    = "// loy:endregion"
)
