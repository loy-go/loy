package architecture_test

import (
	"context"
	"testing"

	"github.com/loy-go/loy/internal/architecture"
	"github.com/loy-go/loy/internal/architecture/rules"
	"github.com/loy-go/loy/internal/filesystem"
	"github.com/loy-go/loy/internal/graph"
)

func TestCustomTopology_CQRS(t *testing.T) {
	customLayers := map[string]architecture.LayerTopologyConfig{
		"api": {
			Allows: []string{"command", "query", "platform"},
			Match:  []string{"internal/api/**"},
		},
		"command": {
			Allows: []string{"database", "event", "platform"},
			Match:  []string{"internal/command/**"},
		},
		"query": {
			Allows: []string{"database", "platform"},
			Match:  []string{"internal/query/**"},
		},
		"database": {
			Allows: []string{"platform"},
			Match:  []string{"internal/database/**"},
		},
	}

	topol, patterns := architecture.BuildTopology("custom", customLayers)
	if !topol.IsCustom() {
		t.Fatalf("expected custom topology")
	}

	// Allowed transitions
	if !topol.IsAllowed("api", "command") {
		t.Errorf("expected api -> command to be allowed")
	}
	if !topol.IsAllowed("api", "query") {
		t.Errorf("expected api -> query to be allowed")
	}
	if !topol.IsAllowed("command", "database") {
		t.Errorf("expected command -> database to be allowed")
	}

	// Disallowed transitions
	if topol.IsAllowed("query", "command") {
		t.Errorf("expected query -> command to be disallowed")
	}
	if topol.IsAllowed("database", "api") {
		t.Errorf("expected database -> api to be disallowed")
	}

	// Test classifier pattern matching
	classifier := architecture.NewClassifier("github.com/example/cqrs", nil)
	classifier.SetPatterns(patterns)

	if lyr := classifier.Classify("github.com/example/cqrs/internal/api/order"); lyr != "api" {
		t.Errorf("expected layer api, got %s", lyr)
	}
	if lyr := classifier.Classify("github.com/example/cqrs/internal/command/create_order"); lyr != "command" {
		t.Errorf("expected layer command, got %s", lyr)
	}
	if lyr := classifier.Classify("github.com/example/cqrs/internal/query/get_order"); lyr != "query" {
		t.Errorf("expected layer query, got %s", lyr)
	}
}

func TestCustomTopology_RuleArch010(t *testing.T) {
	topol := architecture.NewCustomTopology()
	topol.Allow("frontend", "backend")

	g := graph.New()
	g.AddNode("app/frontend", "frontend", []string{"file1.go"})
	g.AddNode("app/backend", "backend", []string{"file2.go"})
	g.AddNode("app/database", "database", []string{"file3.go"})

	// Legal edge: frontend -> backend
	g.AddEdge("app/frontend", "app/backend", "file1.go", 10)
	// Illegal edge: frontend -> database
	g.AddEdge("app/frontend", "app/database", "file2.go", 20)

	analysis := &architecture.Analysis{
		Graph:    g,
		Topology: topol,
	}

	r := &rules.RuleArch010{}
	violations := r.Check(context.Background(), analysis)

	if len(violations) != 1 {
		t.Fatalf("expected 1 violation, got %d: %+v", len(violations), violations)
	}
	if violations[0].RuleID != "ARCH-010" {
		t.Errorf("expected ARCH-010, got %s", violations[0].RuleID)
	}
}

func TestClassifier_WildcardWithRecursiveDoubleStar(t *testing.T) {
	c := architecture.NewClassifier("example.com/proj", nil)
	c.AddPattern("internal/*/service/**", architecture.LayerApplication)
	c.AddPattern("internal/*/repository/**", architecture.LayerInfrastructure)

	layer := c.Classify("example.com/proj/internal/user/service/svc.go")
	if layer != architecture.LayerApplication {
		t.Errorf("expected LayerApplication, got %s", layer)
	}

	layer2 := c.Classify("example.com/proj/internal/order/repository/pg/adapter.go")
	if layer2 != architecture.LayerInfrastructure {
		t.Errorf("expected LayerInfrastructure, got %s", layer2)
	}
}

func TestTopology_CaseInsensitiveAllowed(t *testing.T) {
	top := architecture.NewCustomTopology()
	top.Allow("Domain", "Platform")

	if !top.IsAllowed(architecture.LayerDomain, architecture.LayerPlatform) {
		t.Errorf("expected LayerDomain -> LayerPlatform to be allowed")
	}
	if !top.IsAllowed("domain", "platform") {
		t.Errorf("expected 'domain' -> 'platform' to be allowed")
	}
}

func TestCustomTopology_AnalyzerIntegration(t *testing.T) {
	memFS := filesystem.NewMemFileSystem()
	_ = memFS.MkdirAll("/proj/internal/api", 0755)
	_ = memFS.MkdirAll("/proj/internal/db", 0755)

	apiCode := `package api
import _ "example.com/proj/internal/db"
`
	dbCode := `package db
`
	_ = memFS.WriteFile("/proj/internal/api/api.go", []byte(apiCode), 0644)
	_ = memFS.WriteFile("/proj/internal/db/db.go", []byte(dbCode), 0644)

	// In custom topology, api -> db is disallowed
	topol := architecture.NewCustomTopology()
	patterns := []architecture.PatternRule{
		{Pattern: "internal/api/**", Layer: "api"},
		{Pattern: "internal/db/**", Layer: "db"},
	}

	analyzer := architecture.NewAnalyzer(memFS, architecture.AnalyzerConfig{
		ModuleName: "example.com/proj",
		RootDir:    "/proj",
		Rules:      rules.DefaultRules(),
		Topology:   topol,
		Patterns:   patterns,
	})

	violations, err := analyzer.Run(context.Background())
	if err != nil {
		t.Fatalf("analyzer run failed: %v", err)
	}

	foundViolation := false
	for _, v := range violations {
		if v.RuleID == "ARCH-010" {
			foundViolation = true
			break
		}
	}
	if !foundViolation {
		t.Fatalf("expected ARCH-010 violation for api -> db, got: %+v", violations)
	}
}
