package graph

import (
	"fmt"
	"io"
	"sort"
	"strings"
)

// GraphModel represents the nodes and directed edges of packages with their layers.
type GraphModel struct {
	Packages   map[string]string   // pkg -> layer name
	Edges      map[string][]string // from -> []to
	Violations map[string][]string // from -> []to (violations)
}

// NewGraphModel initializes an empty GraphModel.
func NewGraphModel() *GraphModel {
	return &GraphModel{
		Packages:   make(map[string]string),
		Edges:      make(map[string][]string),
		Violations: make(map[string][]string),
	}
}

// AddPackage registers a package and its layer.
func (gm *GraphModel) AddPackage(pkg, layer string) {
	gm.Packages[pkg] = layer
}

// AddEdge registers an import edge without duplicates.
func (gm *GraphModel) AddEdge(from, to string) {
	for _, existing := range gm.Edges[from] {
		if existing == to {
			return
		}
	}
	gm.Edges[from] = append(gm.Edges[from], to)
}

// AddViolation registers an architectural violation edge without duplicates.
func (gm *GraphModel) AddViolation(from, to string) {
	for _, existing := range gm.Violations[from] {
		if existing == to {
			return
		}
	}
	gm.Violations[from] = append(gm.Violations[from], to)
}

// RenderASCII formats the package hierarchy and imports as an ASCII tree.
func RenderASCII(w io.Writer, gm *GraphModel, violationsOnly bool) {
	pkgs := make([]string, 0, len(gm.Packages))
	for p := range gm.Packages {
		pkgs = append(pkgs, p)
	}
	sort.Strings(pkgs)

	for _, pkg := range pkgs {
		layer := gm.Packages[pkg]
		edges := gm.Edges[pkg]
		violations := gm.Violations[pkg]

		if violationsOnly && len(violations) == 0 {
			continue
		}

		fmt.Fprintf(w, "%s [%s]\n", pkg, layer)

		vSet := make(map[string]bool)
		for _, v := range violations {
			vSet[v] = true
		}

		sort.Strings(edges)
		var displayEdges []string
		for _, target := range edges {
			if vSet[target] || !violationsOnly {
				displayEdges = append(displayEdges, target)
			}
		}

		for i, target := range displayEdges {
			prefix := "├──"
			if i == len(displayEdges)-1 {
				prefix = "└──"
			}
			tLayer := gm.Packages[target]
			if vSet[target] {
				fmt.Fprintf(w, "  %s %s [%s] (VIOLATION)\n", prefix, target, tLayer)
			} else {
				fmt.Fprintf(w, "  %s %s [%s]\n", prefix, target, tLayer)
			}
		}
	}
}

// RenderDOT formats the graph in Graphviz DOT format.
func RenderDOT(w io.Writer, gm *GraphModel, violationsOnly bool) {
	fmt.Fprintln(w, "digraph Architecture {")
	fmt.Fprintln(w, "  rankdir=LR;")
	fmt.Fprintln(w, "  node [shape=box, fontname=\"Helvetica\"];")

	vMap := make(map[string]bool)
	involved := make(map[string]bool)
	for from, targets := range gm.Violations {
		for _, to := range targets {
			vMap[from+"->"+to] = true
			if violationsOnly {
				involved[from] = true
				involved[to] = true
			}
		}
	}

	// Declare nodes with layer colors
	pkgs := make([]string, 0, len(gm.Packages))
	for p := range gm.Packages {
		pkgs = append(pkgs, p)
	}
	sort.Strings(pkgs)

	for _, pkg := range pkgs {
		if violationsOnly && !involved[pkg] {
			continue
		}
		layer := gm.Packages[pkg]
		color := "lightblue"
		switch strings.ToLower(layer) {
		case "domain":
			color = "lightgreen"
		case "application":
			color = "khaki"
		case "infrastructure", "platform":
			color = "lightcoral"
		case "transport":
			color = "plum"
		}
		fmt.Fprintf(w, "  %q [label=\"%s\\n[%s]\", fillcolor=\"%s\", style=\"filled\"];\n", pkg, pkg, layer, color)
	}

	// Declare edges
	for _, from := range pkgs {
		targets := gm.Edges[from]
		sort.Strings(targets)
		for _, to := range targets {
			isViol := vMap[from+"->"+to]
			if violationsOnly && !isViol {
				continue
			}
			if isViol {
				fmt.Fprintf(w, "  %q -> %q [color=\"red\", penwidth=2.0];\n", from, to)
			} else {
				fmt.Fprintf(w, "  %q -> %q;\n", from, to)
			}
		}
	}

	fmt.Fprintln(w, "}")
}

// RenderMermaid formats the graph in Mermaid markdown syntax.
func RenderMermaid(w io.Writer, gm *GraphModel, violationsOnly bool) {
	fmt.Fprintln(w, "```mermaid")
	fmt.Fprintln(w, "graph TD")

	sanitize := func(s string) string {
		r := strings.NewReplacer("/", "_", ".", "_", "-", "_")
		return r.Replace(s)
	}

	vMap := make(map[string]bool)
	involved := make(map[string]bool)
	for from, targets := range gm.Violations {
		for _, to := range targets {
			vMap[from+"->"+to] = true
			if violationsOnly {
				involved[from] = true
				involved[to] = true
			}
		}
	}

	pkgs := make([]string, 0, len(gm.Packages))
	for p := range gm.Packages {
		pkgs = append(pkgs, p)
	}
	sort.Strings(pkgs)

	for _, pkg := range pkgs {
		if violationsOnly && !involved[pkg] {
			continue
		}
		layer := gm.Packages[pkg]
		id := sanitize(pkg)
		fmt.Fprintf(w, "  %s[\"%s<br/>(%s)\"]\n", id, pkg, layer)
	}

	for _, from := range pkgs {
		fromID := sanitize(from)
		targets := gm.Edges[from]
		sort.Strings(targets)
		for _, to := range targets {
			toID := sanitize(to)
			isViol := vMap[from+"->"+to]
			if violationsOnly && !isViol {
				continue
			}
			if isViol {
				fmt.Fprintf(w, "  %s -.->|violation| %s\n", fromID, toID)
			} else {
				fmt.Fprintf(w, "  %s --> %s\n", fromID, toID)
			}
		}
	}

	fmt.Fprintln(w, "```")
}
