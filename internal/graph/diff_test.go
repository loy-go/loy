package graph_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/loy-go/loy/internal/graph"
)

func TestComputeDiff(t *testing.T) {
	base := graph.NewGraphModel()
	base.AddPackage("pkg/domain", "domain")
	base.AddPackage("pkg/service", "application")
	base.AddEdge("pkg/service", "pkg/domain")
	base.AddViolation("pkg/domain", "pkg/service") // old violation

	curr := graph.NewGraphModel()
	curr.AddPackage("pkg/domain", "domain")
	curr.AddPackage("pkg/service", "application")
	curr.AddPackage("pkg/handler", "transport") // newly added package
	curr.AddEdge("pkg/service", "pkg/domain")
	curr.AddEdge("pkg/handler", "pkg/service")  // newly added edge
	curr.AddViolation("pkg/service", "pkg/handler") // newly added violation

	diff := graph.ComputeDiff("origin/main", base, curr)

	if !diff.HasChanges() {
		t.Fatalf("expected changes in diff")
	}

	if len(diff.AddedPackages) != 1 || diff.AddedPackages["pkg/handler"] != "transport" {
		t.Errorf("expected 1 added package pkg/handler, got: %v", diff.AddedPackages)
	}

	if len(diff.AddedEdges) != 1 || diff.AddedEdges[0].From != "pkg/handler" {
		t.Errorf("expected 1 added edge, got: %v", diff.AddedEdges)
	}

	if len(diff.NewViolations) != 1 || diff.NewViolations[0].From != "pkg/service" {
		t.Errorf("expected 1 new violation, got: %v", diff.NewViolations)
	}

	if len(diff.ResolvedViolations) != 1 || diff.ResolvedViolations[0].From != "pkg/domain" {
		t.Errorf("expected 1 resolved violation, got: %v", diff.ResolvedViolations)
	}

	// Test ASCII
	var asciiBuf bytes.Buffer
	graph.RenderDiffASCII(&asciiBuf, diff)
	asciiOut := asciiBuf.String()
	if !strings.Contains(asciiOut, "NEW ARCHITECTURAL VIOLATIONS") || !strings.Contains(asciiOut, "RESOLVED VIOLATIONS") {
		t.Errorf("unexpected ASCII diff output:\n%s", asciiOut)
	}

	// Test Mermaid
	var mBuf bytes.Buffer
	graph.RenderDiffMermaid(&mBuf, diff)
	mOut := mBuf.String()
	if !strings.Contains(mOut, "```mermaid") || !strings.Contains(mOut, "NEW VIOLATION") {
		t.Errorf("unexpected Mermaid diff output:\n%s", mOut)
	}

	// Test Markdown
	var mdBuf bytes.Buffer
	graph.RenderDiffMarkdown(&mdBuf, diff)
	mdOut := mdBuf.String()
	if !strings.Contains(mdOut, "Architecture Drift Report") || !strings.Contains(mdOut, "New Architectural Violation") {
		t.Errorf("unexpected Markdown diff output:\n%s", mdOut)
	}
}

func TestComputeDiff_NoChanges(t *testing.T) {
	base := graph.NewGraphModel()
	base.AddPackage("pkg/domain", "domain")

	curr := graph.NewGraphModel()
	curr.AddPackage("pkg/domain", "domain")

	diff := graph.ComputeDiff("main", base, curr)
	if diff.HasChanges() {
		t.Fatalf("expected no changes")
	}

	var buf bytes.Buffer
	graph.RenderDiffASCII(&buf, diff)
	if !strings.Contains(buf.String(), "No architectural changes detected") {
		t.Errorf("expected no changes message, got: %s", buf.String())
	}

	buf.Reset()
	graph.RenderDiffMarkdown(&buf, diff)
	if !strings.Contains(buf.String(), "No architectural changes detected") {
		t.Errorf("expected clean markdown report, got: %s", buf.String())
	}
}
