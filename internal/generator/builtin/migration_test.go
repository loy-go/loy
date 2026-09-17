package builtin_test

import (
	"context"
	"strings"
	"testing"

	"github.com/loy-go/loy/internal/generator"
	"github.com/loy-go/loy/internal/generator/builtin"
)

func TestMigrationGenerator_Recipes(t *testing.T) {
	gen := builtin.NewMigrationGenerator("github.com/example/app")

	t.Run("raw migration with timeouts", func(t *testing.T) {
		input := generator.Input{
			Name: "add_metadata",
		}
		arts, err := gen.Generate(context.Background(), input)
		if err != nil {
			t.Fatalf("generate raw migration failed: %v", err)
		}
		if len(arts) != 1 {
			t.Fatalf("expected 1 artifact, got %d", len(arts))
		}
		content := string(arts[0].Content)
		if !strings.Contains(content, "SET lock_timeout = '2s';") {
			t.Errorf("expected lock_timeout in raw migration")
		}
	})

	t.Run("index-concurrent recipe", func(t *testing.T) {
		input := generator.Input{
			Name: "idx_users_email",
			Args: map[string]string{
				"recipe": "index-concurrent",
				"table":  "users",
			},
		}
		arts, err := gen.Generate(context.Background(), input)
		if err != nil {
			t.Fatalf("generate index migration failed: %v", err)
		}
		content := string(arts[0].Content)
		if !strings.Contains(content, "-- +goose NO TRANSACTION") {
			t.Errorf("expected NO TRANSACTION in concurrent index migration")
		}
		if !strings.Contains(content, "CREATE INDEX CONCURRENTLY IF NOT EXISTS") {
			t.Errorf("expected CREATE INDEX CONCURRENTLY in output")
		}
	})

	t.Run("shadow-column recipe", func(t *testing.T) {
		input := generator.Input{
			Name: "rename_status",
			Args: map[string]string{
				"recipe": "shadow-column",
				"table":  "orders",
				"type":   "VARCHAR(50)",
			},
		}
		arts, err := gen.Generate(context.Background(), input)
		if err != nil {
			t.Fatalf("generate shadow column migration failed: %v", err)
		}
		content := string(arts[0].Content)
		if !strings.Contains(content, "ALTER TABLE orders ADD COLUMN IF NOT EXISTS") {
			t.Errorf("expected ADD COLUMN in shadow column migration")
		}
		if !strings.Contains(content, "sync_orders_rename_status_shadow") {
			t.Errorf("expected trigger sync function in shadow column migration")
		}
		if strings.Contains(content, "\nUPDATE orders SET") {
			t.Errorf("expected monolithic UPDATE to be omitted from DDL transaction block")
		}
	})

	t.Run("recipe with explicit column", func(t *testing.T) {
		input := generator.Input{
			Name: "add_idx",
			Args: map[string]string{
				"recipe": "index-concurrent",
				"table":  "users",
				"column": "email",
			},
		}
		arts, err := gen.Generate(context.Background(), input)
		if err != nil {
			t.Fatalf("generate failed: %v", err)
		}
		content := string(arts[0].Content)
		if !strings.Contains(content, "ON users (email)") {
			t.Errorf("expected ON users (email), got:\n%s", content)
		}
	})

	t.Run("validation errors on invalid inputs", func(t *testing.T) {
		badTable := generator.Input{
			Name: "test_mig",
			Args: map[string]string{"table": "users; DROP TABLE users;--"},
		}
		if _, err := gen.Generate(context.Background(), badTable); err == nil {
			t.Error("expected error on malicious table name, got nil")
		}

		badColumn := generator.Input{
			Name: "test_mig",
			Args: map[string]string{"column": "col; DROP TABLE users;--"},
		}
		if _, err := gen.Generate(context.Background(), badColumn); err == nil {
			t.Error("expected error on malicious column name, got nil")
		}

		badType := generator.Input{
			Name: "test_mig",
			Args: map[string]string{"type": "TEXT; DROP TABLE users;--"},
		}
		if _, err := gen.Generate(context.Background(), badType); err == nil {
			t.Error("expected error on malicious type, got nil")
		}
	})
}

func TestDualID_Scaffolding(t *testing.T) {
	modelGen := builtin.NewModelGenerator("github.com/example/app")
	input := generator.Input{
		Name: "product",
		Args: map[string]string{
			"dual_id": "true",
			"fields":  "name:string",
		},
	}
	arts, err := modelGen.Generate(context.Background(), input)
	if err != nil {
		t.Fatalf("model generation failed: %v", err)
	}
	content := string(arts[0].Content)
	if !strings.Contains(content, "PublicID") {
		t.Errorf("expected PublicID in dual-id model, got:\n%s", content)
	}
}
