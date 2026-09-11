package workspace

// Workspace models the topology of apps and packages in a Go workspace.
type Workspace struct {
	RootDir  string             `json:"root_dir"`
	Modules  map[string]*Module `json:"modules"`
	Apps     []*Module          `json:"apps"`
	Packages []*Module          `json:"packages"`
}
