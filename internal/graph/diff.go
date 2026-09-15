package graph

import (
	"fmt"
	"io"
	"sort"
)

// EdgePair represents a directed dependency edge between two packages.
type EdgePair struct {
	From string `json:"from"`
	To   string `json:"to"`
}

// DiffModel represents the architectural delta between two graph models.
type DiffModel struct {
	BaseRef            string            `json:"base_ref"`
	AddedPackages      map[string]string `json:"added_packages"`   // pkg -> layer
	RemovedPackages    map[string]string `json:"removed_packages"` // pkg -> layer
	AddedEdges         []EdgePair        `json:"added_edges"`
	RemovedEdges       []EdgePair        `json:"removed_edges"`
	NewViolations      []EdgePair        `json:"new_violations"`
	ResolvedViolations []EdgePair        `json:"resolved_violations"`
}

// ComputeDiff calculates the architectural changes from base to curr.
func ComputeDiff(baseRef string, base, curr *GraphModel) *DiffModel {
	diff := &DiffModel{
		BaseRef:         baseRef,
		AddedPackages:   make(map[string]string),
		RemovedPackages: make(map[string]string),
	}

	if base == nil {
		base = NewGraphModel()
	}
	if curr == nil {
		curr = NewGraphModel()
	}

	// 1. Packages
	for pkg, layer := range curr.Packages {
		if _, exists := base.Packages[pkg]; !exists {
			diff.AddedPackages[pkg] = layer
		}
	}
	for pkg, layer := range base.Packages {
		if _, exists := curr.Packages[pkg]; !exists {
			diff.RemovedPackages[pkg] = layer
		}
	}

	// 2. Edges
	baseEdgeSet := make(map[string]bool)
	for from, targets := range base.Edges {
		for _, to := range targets {
			baseEdgeSet[from+"->"+to] = true
		}
	}

	currEdgeSet := make(map[string]bool)
	for from, targets := range curr.Edges {
		for _, to := range targets {
			currEdgeSet[from+"->"+to] = true
			if !baseEdgeSet[from+"->"+to] {
				diff.AddedEdges = append(diff.AddedEdges, EdgePair{From: from, To: to})
			}
		}
	}

	for from, targets := range base.Edges {
		for _, to := range targets {
			if !currEdgeSet[from+"->"+to] {
				diff.RemovedEdges = append(diff.RemovedEdges, EdgePair{From: from, To: to})
			}
		}
	}

	// 3. Violations
	baseVioSet := make(map[string]bool)
	for from, targets := range base.Violations {
		for _, to := range targets {
			baseVioSet[from+"->"+to] = true
		}
	}

	currVioSet := make(map[string]bool)
	for from, targets := range curr.Violations {
		for _, to := range targets {
			currVioSet[from+"->"+to] = true
			if !baseVioSet[from+"->"+to] {
				diff.NewViolations = append(diff.NewViolations, EdgePair{From: from, To: to})
			}
		}
	}

	for from, targets := range base.Violations {
		for _, to := range targets {
			if !currVioSet[from+"->"+to] {
				diff.ResolvedViolations = append(diff.ResolvedViolations, EdgePair{From: from, To: to})
			}
		}
	}

	// Sort edge slices deterministically
	sortEdges := func(edges []EdgePair) {
		sort.Slice(edges, func(i, j int) bool {
			if edges[i].From == edges[j].From {
				return edges[i].To < edges[j].To
			}
			return edges[i].From < edges[j].From
		})
	}
	sortEdges(diff.AddedEdges)
	sortEdges(diff.RemovedEdges)
	sortEdges(diff.NewViolations)
	sortEdges(diff.ResolvedViolations)

	return diff
}

// HasChanges returns true if any structural changes or violation changes exist.
func (d *DiffModel) HasChanges() bool {
	return len(d.AddedPackages) > 0 ||
		len(d.RemovedPackages) > 0 ||
		len(d.AddedEdges) > 0 ||
		len(d.RemovedEdges) > 0 ||
		len(d.NewViolations) > 0 ||
		len(d.ResolvedViolations) > 0
}

// RenderDiffASCII writes human-readable diff text to w.
func RenderDiffASCII(w io.Writer, diff *DiffModel) {
	_, _ = fmt.Fprintf(w, "Architecture Drift (compared to %s):\n", diff.BaseRef)
	if !diff.HasChanges() {
		_, _ = fmt.Fprintln(w, "  No architectural changes detected.")
		return
	}

	if len(diff.NewViolations) > 0 {
		_, _ = fmt.Fprintln(w, "\n[!] NEW ARCHITECTURAL VIOLATIONS:")
		for _, v := range diff.NewViolations {
			_, _ = fmt.Fprintf(w, "  ! %s -> %s\n", v.From, v.To)
		}
	}

	if len(diff.ResolvedViolations) > 0 {
		_, _ = fmt.Fprintln(w, "\n[✓] RESOLVED VIOLATIONS:")
		for _, v := range diff.ResolvedViolations {
			_, _ = fmt.Fprintf(w, "  ✓ %s -> %s\n", v.From, v.To)
		}
	}

	if len(diff.AddedPackages) > 0 {
		_, _ = fmt.Fprintln(w, "\n[+] Added Packages:")
		var keys []string
		for p := range diff.AddedPackages {
			keys = append(keys, p)
		}
		sort.Strings(keys)
		for _, p := range keys {
			_, _ = fmt.Fprintf(w, "  + %s [%s]\n", p, diff.AddedPackages[p])
		}
	}

	if len(diff.RemovedPackages) > 0 {
		_, _ = fmt.Fprintln(w, "\n[-] Removed Packages:")
		var keys []string
		for p := range diff.RemovedPackages {
			keys = append(keys, p)
		}
		sort.Strings(keys)
		for _, p := range keys {
			_, _ = fmt.Fprintf(w, "  - %s [%s]\n", p, diff.RemovedPackages[p])
		}
	}

	if len(diff.AddedEdges) > 0 {
		_, _ = fmt.Fprintln(w, "\n[+] Added Dependencies:")
		for _, e := range diff.AddedEdges {
			_, _ = fmt.Fprintf(w, "  + %s -> %s\n", e.From, e.To)
		}
	}

	if len(diff.RemovedEdges) > 0 {
		_, _ = fmt.Fprintln(w, "\n[-] Removed Dependencies:")
		for _, e := range diff.RemovedEdges {
			_, _ = fmt.Fprintf(w, "  - %s -> %s\n", e.From, e.To)
		}
	}
}

// RenderDiffMermaid formats the diff as a Mermaid graph with highlighting.
func RenderDiffMermaid(w io.Writer, diff *DiffModel) {
	_, _ = fmt.Fprintln(w, "```mermaid")
	_, _ = fmt.Fprintln(w, "graph TD")

	nodeID := func(pkg string) string {
		var safe []rune
		for _, r := range pkg {
			if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
				safe = append(safe, r)
			} else {
				safe = append(safe, '_')
			}
		}
		return string(safe)
	}

	for _, e := range diff.AddedEdges {
		fromID := nodeID(e.From)
		toID := nodeID(e.To)
		_, _ = fmt.Fprintf(w, "  %s([%s]) -->|added| %s([%s])\n", fromID, e.From, toID, e.To)
	}

	for _, v := range diff.NewViolations {
		fromID := nodeID(v.From)
		toID := nodeID(v.To)
		_, _ = fmt.Fprintf(w, "  %s -.->|NEW VIOLATION| %s\n", fromID, toID)
		_, _ = fmt.Fprintf(w, "  linkStyle default stroke:red,stroke-width:2px;\n")
	}

	_, _ = fmt.Fprintln(w, "```")
}

// RenderDiffMarkdown formats a Markdown report suitable for PR comments.
func RenderDiffMarkdown(w io.Writer, diff *DiffModel) {
	_, _ = fmt.Fprintf(w, "### 🏛️ Architecture Drift Report\n\n")
	_, _ = fmt.Fprintf(w, "**Comparison Base:** `%s`\n\n", diff.BaseRef)

	if !diff.HasChanges() {
		_, _ = fmt.Fprintln(w, "✅ **No architectural changes detected.** Package dependencies and layer boundaries are unchanged.")
		return
	}

	if len(diff.NewViolations) > 0 {
		_, _ = fmt.Fprintf(w, "🚨 **%d New Architectural Violation(s) Introduced!**\n\n", len(diff.NewViolations))
		_, _ = fmt.Fprintln(w, "| From Package | To Package | Status |")
		_, _ = fmt.Fprintln(w, "|---|---|---|")
		for _, v := range diff.NewViolations {
			_, _ = fmt.Fprintf(w, "| `%s` | `%s` | ❌ VIOLATION |\n", v.From, v.To)
		}
		_, _ = fmt.Fprintln(w, "")
	} else {
		_, _ = fmt.Fprintln(w, "✅ **No new architectural violations introduced.**")
		_, _ = fmt.Fprintln(w, "")
	}

	if len(diff.ResolvedViolations) > 0 {
		_, _ = fmt.Fprintf(w, "🎉 **%d Violation(s) Resolved:**\n\n", len(diff.ResolvedViolations))
		for _, v := range diff.ResolvedViolations {
			_, _ = fmt.Fprintf(w, "- `%s` ➔ `%s`\n", v.From, v.To)
		}
		_, _ = fmt.Fprintln(w, "")
	}

	if len(diff.AddedPackages) > 0 || len(diff.RemovedPackages) > 0 {
		_, _ = fmt.Fprintln(w, "<details><summary>📦 Package Changes</summary>")
		_, _ = fmt.Fprintln(w, "")
		if len(diff.AddedPackages) > 0 {
			_, _ = fmt.Fprintln(w, "**Added:**")
			var addedKeys []string
			for p := range diff.AddedPackages {
				addedKeys = append(addedKeys, p)
			}
			sort.Strings(addedKeys)
			for _, p := range addedKeys {
				_, _ = fmt.Fprintf(w, "- `+ %s` (%s)\n", p, diff.AddedPackages[p])
			}
		}
		if len(diff.RemovedPackages) > 0 {
			_, _ = fmt.Fprintln(w, "**Removed:**")
			var removedKeys []string
			for p := range diff.RemovedPackages {
				removedKeys = append(removedKeys, p)
			}
			sort.Strings(removedKeys)
			for _, p := range removedKeys {
				_, _ = fmt.Fprintf(w, "- `- %s` (%s)\n", p, diff.RemovedPackages[p])
			}
		}
		_, _ = fmt.Fprintln(w, "</details>")
	}
}
