package builtin_test

import (
	"context"
	"strings"
	"testing"

	"github.com/uloydev/loy/internal/generator"
	"github.com/uloydev/loy/internal/generator/builtin"
)

func TestViewGenerator(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	mod := "github.com/example/fullstackapp"

	t.Run("full page view", func(t *testing.T) {
		gen := builtin.NewViewGenerator(mod)
		if gen.Name() != "view" {
			t.Fatalf("expected view generator name, got %s", gen.Name())
		}

		artifacts, err := gen.Generate(ctx, generator.Input{
			Name: "home",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(artifacts) != 1 {
			t.Fatalf("expected 1 artifact, got %d", len(artifacts))
		}
		art := artifacts[0]
		if art.Path != "views/pages/home.templ" {
			t.Fatalf("expected path views/pages/home.templ, got %s", art.Path)
		}
		content := string(art.Content)
		if !strings.Contains(content, "package pages") {
			t.Errorf("missing package pages:\n%s", content)
		}
		if !strings.Contains(content, "templ Home()") {
			t.Errorf("missing templ Home():\n%s", content)
		}
		if !strings.Contains(content, "github.com/example/fullstackapp/views/layouts") {
			t.Errorf("missing layouts import:\n%s", content)
		}
	})

	t.Run("partial component view", func(t *testing.T) {
		gen := builtin.NewViewGenerator(mod)
		artifacts, err := gen.Generate(ctx, generator.Input{
			Name: "user_card",
			Args: map[string]string{
				"partial": "true",
			},
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(artifacts) != 1 {
			t.Fatalf("expected 1 artifact, got %d", len(artifacts))
		}
		art := artifacts[0]
		if art.Path != "views/components/user_card.templ" {
			t.Fatalf("expected views/components/user_card.templ, got %s", art.Path)
		}
		content := string(art.Content)
		if !strings.Contains(content, "package components") {
			t.Errorf("missing package components:\n%s", content)
		}
		if !strings.Contains(content, "templ UserCard()") {
			t.Errorf("missing templ UserCard():\n%s", content)
		}
		if !strings.Contains(content, "hx-target") {
			t.Errorf("missing hx-target attribute:\n%s", content)
		}
	})

	t.Run("nested subpaths", func(t *testing.T) {
		gen := builtin.NewViewGenerator(mod)
		artifacts, err := gen.Generate(ctx, generator.Input{
			Name: "admin/dashboard",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(artifacts) != 1 {
			t.Fatalf("expected 1 artifact, got %d", len(artifacts))
		}
		art := artifacts[0]
		if art.Path != "views/pages/admin/dashboard.templ" {
			t.Fatalf("expected views/pages/admin/dashboard.templ, got %s", art.Path)
		}
		content := string(art.Content)
		if !strings.Contains(content, "package admin") {
			t.Errorf("missing package admin:\n%s", content)
		}
		if !strings.Contains(content, "templ Dashboard()") {
			t.Errorf("missing templ Dashboard():\n%s", content)
		}
	})

	t.Run("nested partial component", func(t *testing.T) {
		gen := builtin.NewViewGenerator(mod)
		artifacts, err := gen.Generate(ctx, generator.Input{
			Name: "admin/stats",
			Args: map[string]string{
				"partial": "true",
			},
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(artifacts) != 1 {
			t.Fatalf("expected 1 artifact, got %d", len(artifacts))
		}
		art := artifacts[0]
		if art.Path != "views/components/admin/stats.templ" {
			t.Fatalf("expected views/components/admin/stats.templ, got %s", art.Path)
		}
		content := string(art.Content)
		if !strings.Contains(content, "package admin") {
			t.Errorf("missing package admin:\n%s", content)
		}
	})

	t.Run("validation errors", func(t *testing.T) {
		gen := builtin.NewViewGenerator(mod)

		if _, err := gen.Generate(ctx, generator.Input{Name: ""}); err == nil {
			t.Error("expected error on empty name")
		}

		if _, err := gen.Generate(ctx, generator.Input{Name: "../outside"}); err == nil {
			t.Error("expected error on path traversal")
		}
	})
}
