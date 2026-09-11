package cli_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/uloydev/loy/internal/cli"
)

func TestMakeCRUDAndJSON(t *testing.T) {
	tempDir := t.TempDir()
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd error: %v", err)
	}
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("chdir error: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(origDir)
	})

	if err := os.WriteFile("go.mod", []byte("module github.com/example/shop\n\ngo 1.22\n"), 0644); err != nil {
		t.Fatalf("writing go.mod error: %v", err)
	}

	root := cli.NewRootCmd()

	t.Run("make crud with json output", func(t *testing.T) {
		out, err := executeMakeCommand(root, "make", "crud", "product", "title:string:unique", "price:float", "--json")
		if err != nil {
			t.Fatalf("make crud failed: %v, out: %s", err, out)
		}

		// Verify files generated
		migs, _ := filepath.Glob(filepath.Join(tempDir, "migrations/*_create_products_table.sql"))
		if len(migs) != 1 {
			t.Fatalf("expected 1 migration file, found %d", len(migs))
		}

		queryPath := filepath.Join(tempDir, "queries/products.sql")
		if _, err := os.Stat(queryPath); os.IsNotExist(err) {
			t.Fatalf("expected query file %s to exist", queryPath)
		}

		modelPath := filepath.Join(tempDir, "internal/product/domain/product.go")
		if _, err := os.Stat(modelPath); os.IsNotExist(err) {
			t.Fatalf("expected model file %s to exist", modelPath)
		}
	})

	t.Run("make crud without force fails on developer owned existing files", func(t *testing.T) {
		_, err := executeMakeCommand(root, "make", "crud", "product", "title:string")
		if err == nil {
			t.Fatalf("expected error on conflict without force, got nil")
		}
	})

	t.Run("make crud with force overwrites successfully", func(t *testing.T) {
		out, err := executeMakeCommand(root, "make", "crud", "product", "title:string", "--force")
		if err != nil {
			t.Fatalf("expected make crud --force to succeed, got %v (out: %s)", err, out)
		}
	})
}
