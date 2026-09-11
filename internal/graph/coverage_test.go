package graph

import (
	"reflect"
	"testing"
)

func TestGraphCoverage(t *testing.T) {
	g := New()
	n1 := g.AddNode("A", "Transport", []string{"a.go"})
	if n1.Layer != "Transport" {
		t.Errorf("expected Transport, got %s", n1.Layer)
	}

	// Update existing node
	n1Updated := g.AddNode("A", "", []string{"a2.go"})
	if len(n1Updated.FilePaths) != 2 {
		t.Errorf("expected 2 filepaths, got %d", len(n1Updated.FilePaths))
	}

	g.AddEdge("A", "B", "a.go", 1)
	edges := g.AllEdges()
	if len(edges) != 1 {
		t.Fatalf("expected 1 edge, got %d", len(edges))
	}

	nodes := g.Nodes()
	if len(nodes) != 2 {
		t.Fatalf("expected 2 nodes (A and B auto-created), got %d", len(nodes))
	}

	// Empty cycle reconstruction check
	cycle := g.reconstructCycle([]string{})
	if len(cycle.Path) != 0 {
		t.Errorf("expected empty path, got %+v", cycle.Path)
	}

	// 2-node cycle reconstruction
	g2 := New()
	g2.AddNode("X", "", nil)
	g2.AddNode("Y", "", nil)
	g2.AddEdge("X", "Y", "x.go", 1)
	g2.AddEdge("Y", "X", "y.go", 1)
	c2 := g2.reconstructCycle([]string{"X", "Y"})
	expected := []string{"X", "Y", "X"}
	if !reflect.DeepEqual(c2.Path, expected) {
		t.Errorf("expected %+v, got %+v", expected, c2.Path)
	}
}
