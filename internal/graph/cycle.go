package graph

import (
	"sort"
)

// Cycle represents a detected cyclic dependency path.
type Cycle struct {
	Path []string // e.g. ["A", "B", "C", "A"]
}

// FindCycles uses Tarjan's strongly connected components algorithm
// to detect all cycles with length >= 2 or self-referential cycles.
func (g *Graph) FindCycles() []Cycle {
	var (
		index     int
		indices   = make(map[string]int)
		lowlink   = make(map[string]int)
		onStack   = make(map[string]bool)
		stack     []string
		sccs      [][]string
	)

	// Collect and sort node keys for determinism
	var nodeKeys []string
	for k := range g.nodes {
		nodeKeys = append(nodeKeys, k)
	}
	sort.Strings(nodeKeys)

	var strongConnect func(v string)
	strongConnect = func(v string) {
		indices[v] = index
		lowlink[v] = index
		index++
		stack = append(stack, v)
		onStack[v] = true

		for _, edge := range g.EdgesFrom(v) {
			w := edge.To
			// Only consider nodes actually in graph nodes set
			if _, exists := g.nodes[w]; !exists {
				continue
			}
			if _, visited := indices[w]; !visited {
				strongConnect(w)
				if lowlink[w] < lowlink[v] {
					lowlink[v] = lowlink[w]
				}
			} else if onStack[w] {
				if indices[w] < lowlink[v] {
					lowlink[v] = indices[w]
				}
			}
		}

		if lowlink[v] == indices[v] {
			var scc []string
			for {
				w := stack[len(stack)-1]
				stack = stack[:len(stack)-1]
				onStack[w] = false
				scc = append(scc, w)
				if w == v {
					break
				}
			}
			// An SCC is a cycle if size > 1 OR size == 1 with self-edge
			if len(scc) > 1 {
				sccs = append(sccs, scc)
			} else if len(scc) == 1 {
				u := scc[0]
				for _, edge := range g.EdgesFrom(u) {
					if edge.To == u {
						sccs = append(sccs, scc)
						break
					}
				}
			}
		}
	}

	for _, v := range nodeKeys {
		if _, visited := indices[v]; !visited {
			strongConnect(v)
		}
	}

	var cycles []Cycle
	for _, scc := range sccs {
		cycles = append(cycles, g.reconstructCycle(scc))
	}
	return cycles
}

// reconstructCycle orders an SCC into an explicit chain path.
func (g *Graph) reconstructCycle(scc []string) Cycle {
	if len(scc) == 0 {
		return Cycle{}
	}
	if len(scc) == 1 {
		return Cycle{Path: []string{scc[0], scc[0]}}
	}

	sccSet := make(map[string]bool)
	for _, s := range scc {
		sccSet[s] = true
	}

	// Pick lowest lexical start
	start := scc[0]
	for _, s := range scc {
		if s < start {
			start = s
		}
	}

	visited := make(map[string]bool)
	var path []string
	var dfs func(curr string) bool
	dfs = func(curr string) bool {
		path = append(path, curr)
		visited[curr] = true

		for _, edge := range g.EdgesFrom(curr) {
			if !sccSet[edge.To] {
				continue
			}
			if edge.To == start && len(path) > 1 {
				path = append(path, start)
				return true
			}
			if !visited[edge.To] {
				if dfs(edge.To) {
					return true
				}
			}
		}
		path = path[:len(path)-1]
		visited[curr] = false
		return false
	}

	if dfs(start) {
		return Cycle{Path: path}
	}

	// Fallback if simple DFS loop fails to close cleanly
	fallback := append([]string{}, scc...)
	sort.Strings(fallback)
	fallback = append(fallback, fallback[0])
	return Cycle{Path: fallback}
}
