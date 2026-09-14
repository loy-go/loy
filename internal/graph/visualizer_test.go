package graph_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/loy-go/loy/internal/graph"
)

func TestVisualizerRenderers(t *testing.T) {
	gm := graph.NewGraphModel()
	gm.AddPackage("app/domain", "Domain")
	gm.AddPackage("app/service", "Application")
	gm.AddPackage("app/transport/http", "Transport")

	gm.AddEdge("app/service", "app/domain")
	gm.AddEdge("app/transport/http", "app/service")
	gm.AddEdge("app/transport/http", "app/domain")
	gm.AddViolation("app/transport/http", "app/domain") // direct violation

	// 1. ASCII
	asciiBuf := new(bytes.Buffer)
	graph.RenderASCII(asciiBuf, gm, false)
	asciiOut := asciiBuf.String()
	if !strings.Contains(asciiOut, "app/domain [Domain]") || !strings.Contains(asciiOut, "(VIOLATION)") {
		t.Errorf("unexpected ASCII output:\n%s", asciiOut)
	}

	// 2. DOT
	dotBuf := new(bytes.Buffer)
	graph.RenderDOT(dotBuf, gm, false)
	dotOut := dotBuf.String()
	if !strings.Contains(dotOut, "digraph Architecture") || !strings.Contains(dotOut, "color=\"red\"") {
		t.Errorf("unexpected DOT output:\n%s", dotOut)
	}

	// 3. Mermaid
	mBuf := new(bytes.Buffer)
	graph.RenderMermaid(mBuf, gm, false)
	mOut := mBuf.String()
	if !strings.Contains(mOut, "```mermaid") || !strings.Contains(mOut, "graph TD") || !strings.Contains(mOut, "-.->|violation|") {
		t.Errorf("unexpected Mermaid output:\n%s", mOut)
	}
}
