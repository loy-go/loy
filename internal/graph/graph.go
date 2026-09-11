package graph

// Edge represents a directed dependency edge between packages or modules.
type Edge struct {
	From string
	To   string
	File string
	Line int
}

// Node represents a package in the dependency graph.
type Node struct {
	Path      string
	Layer     string
	FilePaths []string
	Metadata  map[string]any
}

// Graph is a directed graph of package dependencies.
type Graph struct {
	nodes map[string]*Node
	edges map[string][]Edge // key is From
}

// New creates an empty directed graph.
func New() *Graph {
	return &Graph{
		nodes: make(map[string]*Node),
		edges: make(map[string][]Edge),
	}
}

// AddNode registers a node if not already present.
func (g *Graph) AddNode(path string, layer string, filePaths []string) *Node {
	if n, ok := g.nodes[path]; ok {
		if layer != "" && n.Layer == "" {
			n.Layer = layer
		}
		if len(filePaths) > 0 {
			n.FilePaths = append(n.FilePaths, filePaths...)
		}
		return n
	}
	n := &Node{
		Path:      path,
		Layer:     layer,
		FilePaths: filePaths,
		Metadata:  make(map[string]any),
	}
	g.nodes[path] = n
	return n
}

// Node returns a node by path or nil.
func (g *Graph) Node(path string) *Node {
	return g.nodes[path]
}

// Nodes returns all nodes in graph.
func (g *Graph) Nodes() map[string]*Node {
	return g.nodes
}

// AddEdge registers a directed edge from -> to.
func (g *Graph) AddEdge(from, to, file string, line int) {
	if _, ok := g.nodes[from]; !ok {
		g.AddNode(from, "", nil)
	}
	if _, ok := g.nodes[to]; !ok {
		g.AddNode(to, "", nil)
	}
	g.edges[from] = append(g.edges[from], Edge{
		From: from,
		To:   to,
		File: file,
		Line: line,
	})
}

// EdgesFrom returns all outbound edges from given node path.
func (g *Graph) EdgesFrom(from string) []Edge {
	return g.edges[from]
}

// AllEdges returns all edges across the graph.
func (g *Graph) AllEdges() []Edge {
	var all []Edge
	for _, es := range g.edges {
		all = append(all, es...)
	}
	return all
}
