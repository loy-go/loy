package builtin_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/loy-go/loy/internal/generator"
	"github.com/loy-go/loy/internal/generator/builtin"
)

func TestGoldenGenerators(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	mod := "github.com/example/myapp"

	goldenDir := filepath.Join("..", "..", "..", "testdata", "golden", "generators")
	_ = os.MkdirAll(goldenDir, 0755)

	normalize := func(s string) string {
		return strings.ReplaceAll(s, "\r\n", "\n")
	}

	t.Run("model golden comparison", func(t *testing.T) {
		gen := builtin.NewModelGenerator(mod)
		arts, err := gen.Generate(ctx, generator.Input{
			Name: "user",
			Args: map[string]string{
				"fields": "email:string:unique age:int:optional",
			},
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(arts) == 0 {
			t.Fatalf("expected artifacts, got none")
		}

		goldenFile := filepath.Join(goldenDir, "user_model.go.golden")
		actual := string(arts[0].Content)

		if os.Getenv("UPDATE_GOLDEN") == "1" {
			_ = os.WriteFile(goldenFile, arts[0].Content, 0644)
		}

		if data, err := os.ReadFile(goldenFile); err == nil {
			if normalize(string(data)) != normalize(actual) {
				t.Fatalf("golden file mismatch: want %s, got %s", string(data), actual)
			}
		} else {
			_ = os.WriteFile(goldenFile, arts[0].Content, 0644)
		}
	})

	t.Run("service golden comparison", func(t *testing.T) {
		gen := builtin.NewServiceGenerator(mod)
		arts, err := gen.Generate(ctx, generator.Input{
			Name: "user",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(arts) == 0 {
			t.Fatalf("expected artifacts, got none")
		}

		goldenFile := filepath.Join(goldenDir, "user_service.go.golden")
		actual := string(arts[0].Content)

		if os.Getenv("UPDATE_GOLDEN") == "1" {
			_ = os.WriteFile(goldenFile, arts[0].Content, 0644)
		}

		if data, err := os.ReadFile(goldenFile); err == nil {
			if normalize(string(data)) != normalize(actual) {
				t.Fatalf("golden file mismatch: want %s, got %s", string(data), actual)
			}
		} else {
			_ = os.WriteFile(goldenFile, arts[0].Content, 0644)
		}
	})
}
