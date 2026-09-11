package graph

import (
	"reflect"
	"testing"
)

func TestGraphCycleDetection(t *testing.T) {
	t.Run("no cycles in DAG", func(t *testing.T) {
		g := New()
		g.AddNode("A", "Transport", nil)
		g.AddNode("B", "Application", nil)
		g.AddNode("C", "Domain", nil)

		g.AddEdge("A", "B", "a.go", 5)
		g.AddEdge("B", "C", "b.go", 10)

		cycles := g.FindCycles()
		if len(cycles) != 0 {
			t.Fatalf("expected 0 cycles, got %d: %+v", len(cycles), cycles)
		}
	})

	t.Run("simple 2-node cycle", func(t *testing.T) {
		g := New()
		g.AddNode("pkg/a", "", nil)
		g.AddNode("pkg/b", "", nil)

		g.AddEdge("pkg/a", "pkg/b", "a.go", 1)
		g.AddEdge("pkg/b", "pkg/a", "b.go", 2)

		cycles := g.FindCycles()
		if len(cycles) != 1 {
			t.Fatalf("expected 1 cycle, got %d: %+v", len(cycles), cycles)
		}
		expected := []string{"pkg/a", "pkg/b", "pkg/a"}
		if !reflect.DeepEqual(cycles[0].Path, expected) {
			t.Errorf("expected cycle %+v, got %+v", expected, cycles[0].Path)
		}
	})

	t.Run("3-node cycle with branch", func(t *testing.T) {
		g := New()
		g.AddNode("A", "", nil)
		g.AddNode("B", "", nil)
		g.AddNode("C", "", nil)
		g.AddNode("D", "", nil)

		g.AddEdge("A", "B", "", 0)
		g.AddEdge("B", "C", "", 0)
		g.AddEdge("C", "A", "", 0)
		g.AddEdge("C", "D", "", 0) // branch out

		cycles := g.FindCycles()
		if len(cycles) != 1 {
			t.Fatalf("expected 1 cycle, got %d", len(cycles))
		}
		expected := []string{"A", "B", "C", "A"}
		if !reflect.DeepEqual(cycles[0].Path, expected) {
			t.Errorf("expected %+v, got %+v", expected, cycles[0].Path)
		}
	})

	t.Run("self-cycle", func(t *testing.T) {
		g := New()
		g.AddNode("A", "", nil)
		g.AddEdge("A", "A", "a.go", 1)

		cycles := g.FindCycles()
		if len(cycles) != 1 {
			t.Fatalf("expected 1 cycle, got %d", len(cycles))
		}
		expected := []string{"A", "A"}
		if !reflect.DeepEqual(cycles[0].Path, expected) {
			t.Errorf("expected %+v, got %+v", expected, cycles[0].Path)
		}
	})
}
