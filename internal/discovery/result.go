package discovery

// DiscoveredContext represents the detected project and workspace structure.
type DiscoveredContext struct {
	RootDir      string
	ManifestPath string
	HasManifest  bool
	HasGoMod     bool
	GoModPath    string
	HasGoWork    bool
	GoWorkPath   string
	IsWorkspace  bool
}
