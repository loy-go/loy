package builtin_test

import (
	"context"
	"strings"
	"testing"

	"github.com/loy-go/loy/internal/generator"
	"github.com/loy-go/loy/internal/generator/builtin"
)

func TestCIGenerator(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	mod := "github.com/example/demoapp"

	t.Run("default github actions", func(t *testing.T) {
		gen := builtin.NewCIGenerator(mod).WithCapabilities(true, true, true)
		artifacts, err := gen.Generate(ctx, generator.Input{Name: "demoapp"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(artifacts) != 1 {
			t.Fatalf("expected 1 artifact, got %d", len(artifacts))
		}
		if artifacts[0].Path != ".github/workflows/ci.yml" {
			t.Errorf("expected .github/workflows/ci.yml, got %s", artifacts[0].Path)
		}

		content := string(artifacts[0].Content)
		if !strings.Contains(content, "uses: actions/setup-go@v5") {
			t.Errorf("missing setup-go in github actions:\n%s", content)
		}
		if !strings.Contains(content, "cache: true") {
			t.Errorf("missing cache: true in github actions:\n%s", content)
		}
		if !strings.Contains(content, "go run github.com/loy-go/loy/cmd/loy@latest check") {
			t.Errorf("missing loy check in github actions:\n%s", content)
		}
		if !strings.Contains(content, "image: postgres:16-alpine") {
			t.Errorf("missing postgres service in github actions:\n%s", content)
		}
	})

	t.Run("gitlab ci provider", func(t *testing.T) {
		gen := builtin.NewCIGenerator(mod).WithProvider("gitlab")
		artifacts, err := gen.Generate(ctx, generator.Input{
			Name: "demoapp",
			Args: map[string]string{
				"provider": "gitlab",
			},
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(artifacts) != 1 {
			t.Fatalf("expected 1 artifact, got %d", len(artifacts))
		}
		if artifacts[0].Path != ".gitlab-ci.yml" {
			t.Errorf("expected .gitlab-ci.yml, got %s", artifacts[0].Path)
		}

		content := string(artifacts[0].Content)
		if !strings.Contains(content, "image: golang:1.23-alpine") {
			t.Errorf("missing golang image in gitlab ci:\n%s", content)
		}
	})
}
