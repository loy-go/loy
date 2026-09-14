package builtin_test

import (
	"context"
	"testing"

	"github.com/uloydev/loy/internal/generator"
	"github.com/uloydev/loy/internal/generator/builtin"
)

func TestGeneratorErrors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	t.Run("empty names error", func(t *testing.T) {
		gens := []generator.Generator{
			builtin.NewModelGenerator("mod"),
			builtin.NewRepositoryGenerator("mod"),
			builtin.NewServiceGenerator("mod"),
			builtin.NewHandlerGenerator("mod"),
			builtin.NewRequestGenerator("mod"),
			builtin.NewResourceGenerator("mod"),
			builtin.NewJobGenerator("mod"),
			builtin.NewEventGenerator("mod"),
			builtin.NewListenerGenerator("mod"),
			builtin.NewPolicyGenerator("mod"),
			builtin.NewTestGenerator("mod"),
			builtin.NewFeatureGenerator("mod"),
			builtin.NewCRUDGenerator("mod"),
		}

		for _, g := range gens {
			_, err := g.Generate(ctx, generator.Input{})
			if err == nil {
				t.Fatalf("expected error for empty name in %s, got nil", g.Name())
			}
		}
	})

	t.Run("invalid field parsing errors in model, request, resource, crud", func(t *testing.T) {
		badInput := generator.Input{
			Name: "user",
			Args: map[string]string{"fields": "bad_field_without_colon"},
		}

		modelGen := builtin.NewModelGenerator("mod")
		if _, err := modelGen.Generate(ctx, badInput); err == nil {
			t.Fatalf("expected error from modelGen")
		}

		reqGen := builtin.NewRequestGenerator("mod")
		if _, err := reqGen.Generate(ctx, badInput); err == nil {
			t.Fatalf("expected error from reqGen")
		}

		resGen := builtin.NewResourceGenerator("mod")
		if _, err := resGen.Generate(ctx, badInput); err == nil {
			t.Fatalf("expected error from resGen")
		}

		crudGen := builtin.NewCRUDGenerator("mod")
		if _, err := crudGen.Generate(ctx, badInput); err == nil {
			t.Fatalf("expected error from crudGen")
		}
	})

	t.Run("malicious entity names rejected", func(t *testing.T) {
		badNames := []string{
			"user; DROP TABLE users;--",
			"bad name with spaces",
			"123startwithnumber",
			"foo-bar-dash",
		}

		gens := []generator.Generator{
			builtin.NewModelGenerator("mod"),
			builtin.NewFeatureGenerator("mod"),
			builtin.NewCRUDGenerator("mod"),
		}

		for _, g := range gens {
			for _, name := range badNames {
				_, err := g.Generate(ctx, generator.Input{Name: name})
				if err == nil {
					t.Fatalf("expected error for malicious entity name %q in %s, got nil", name, g.Name())
				}
			}
		}
	})
}
