package builtin_test

import (
	"context"
	"strings"
	"testing"

	"github.com/loy-go/loy/internal/generator"
	"github.com/loy-go/loy/internal/generator/builtin"
)

func TestAgentRulesGenerator(t *testing.T) {
	ctx := context.Background()

	t.Run("generate all target produces all rule artifacts", func(t *testing.T) {
		gen := builtin.NewAgentRulesGenerator("github.com/example/shop")
		artifacts, err := gen.Generate(ctx, generator.Input{Name: "shop"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		expectedFiles := []string{
			"AGENTS.md",
			"CLAUDE.md",
			".cursor/rules/loy.mdc",
			".cursorrules",
			".github/copilot-instructions.md",
			".windsurfrules",
		}

		if len(artifacts) != len(expectedFiles) {
			t.Fatalf("expected %d artifacts, got %d", len(expectedFiles), len(artifacts))
		}

		found := make(map[string]string)
		for _, a := range artifacts {
			found[a.Path] = string(a.Content)
		}

		for _, f := range expectedFiles {
			content, exists := found[f]
			if !exists {
				t.Errorf("expected artifact %s to be generated", f)
				continue
			}
			if !strings.Contains(content, "Clean Architecture") {
				t.Errorf("expected %s to mention Clean Architecture", f)
			}
		}

		// Verify AGENTS.md mentions loy check --format agent
		if !strings.Contains(found["AGENTS.md"], "loy check --format agent") {
			t.Errorf("expected AGENTS.md to mention loy check --format agent")
		}
	})

	t.Run("generate cursor target produces only cursor rules", func(t *testing.T) {
		gen := builtin.NewAgentRulesGenerator("github.com/example/shop").WithTarget("cursor")
		artifacts, err := gen.Generate(ctx, generator.Input{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(artifacts) != 2 {
			t.Fatalf("expected 2 cursor artifacts, got %d", len(artifacts))
		}
		for _, a := range artifacts {
			if !strings.Contains(a.Path, "cursor") {
				t.Errorf("unexpected artifact for cursor target: %s", a.Path)
			}
		}
	})

	t.Run("generate claude target produces AGENTS.md and CLAUDE.md", func(t *testing.T) {
		gen := builtin.NewAgentRulesGenerator("github.com/example/shop").WithTarget("claude")
		artifacts, err := gen.Generate(ctx, generator.Input{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(artifacts) != 2 {
			t.Fatalf("expected 2 claude artifacts, got %d", len(artifacts))
		}
	})

	t.Run("generate copilot target produces copilot instructions", func(t *testing.T) {
		gen := builtin.NewAgentRulesGenerator("github.com/example/shop").WithTarget("copilot")
		artifacts, err := gen.Generate(ctx, generator.Input{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(artifacts) != 1 || artifacts[0].Path != ".github/copilot-instructions.md" {
			t.Fatalf("expected 1 copilot artifact, got: %+v", artifacts)
		}
	})

	t.Run("generate windsurf target produces windsurfrules", func(t *testing.T) {
		gen := builtin.NewAgentRulesGenerator("github.com/example/shop").WithTarget("windsurf")
		artifacts, err := gen.Generate(ctx, generator.Input{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(artifacts) != 1 || artifacts[0].Path != ".windsurfrules" {
			t.Fatalf("expected 1 windsurf artifact, got: %+v", artifacts)
		}
	})
}
