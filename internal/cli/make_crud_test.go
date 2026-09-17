package cli_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/loy-go/loy/internal/cli"
	"github.com/loy-go/loy/internal/filesystem"
	"github.com/loy-go/loy/internal/process"
)

func newCRUDTestRootCmd() *cobra.Command {
	return cli.NewRootCmdWithFS(filesystem.NewOSFileSystem(), process.NewNoopRunner())
}

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

	root := newCRUDTestRootCmd()

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

func TestMakeCRUD_WithChiAndNetHTTP(t *testing.T) {
	t.Run("chi handler generation", func(t *testing.T) {
		tempDir := t.TempDir()
		origDir, _ := os.Getwd()
		_ = os.Chdir(tempDir)
		t.Cleanup(func() { _ = os.Chdir(origDir) })

		_ = os.WriteFile("go.mod", []byte("module github.com/example/chi-app\n\ngo 1.22\n"), 0644)
		_ = os.WriteFile("loy.yaml", []byte("version: 1\nproject:\n  name: chi-app\ndefaults:\n  http: chi\n"), 0644)

		root := newCRUDTestRootCmd()
		out, err := executeMakeCommand(root, "make", "crud", "order", "code:string")
		if err != nil {
			t.Fatalf("make crud in chi project failed: %v, out: %s", err, out)
		}

		handlerBytes, err := os.ReadFile(filepath.Join(tempDir, "internal/order/transport/http/handler.go"))
		if err != nil {
			t.Fatalf("reading order handler: %v", err)
		}
		if !strings.Contains(string(handlerBytes), "RegisterRoutes(router chi.Router)") {
			t.Errorf("expected chi.Router in generated handler:\n%s", string(handlerBytes))
		}

		// Verify route discovery
		routesOut, routesErr := executeMakeCommand(root, "routes")
		if routesErr != nil {
			t.Fatalf("routes command failed: %v, out: %s", routesErr, routesOut)
		}
		if !strings.Contains(routesOut, "GET") {
			t.Errorf("expected routes in routes output: %s", routesOut)
		}
	})

	t.Run("nethttp handler generation", func(t *testing.T) {
		tempDir := t.TempDir()
		origDir, _ := os.Getwd()
		_ = os.Chdir(tempDir)
		t.Cleanup(func() { _ = os.Chdir(origDir) })

		_ = os.WriteFile("go.mod", []byte("module github.com/example/net-app\n\ngo 1.22\n"), 0644)
		_ = os.WriteFile("loy.yaml", []byte("version: 1\nproject:\n  name: net-app\ndefaults:\n  http: nethttp\n"), 0644)

		root := newCRUDTestRootCmd()
		out, err := executeMakeCommand(root, "make", "crud", "customer", "name:string")
		if err != nil {
			t.Fatalf("make crud in nethttp project failed: %v, out: %s", err, out)
		}

		handlerBytes, err := os.ReadFile(filepath.Join(tempDir, "internal/customer/transport/http/handler.go"))
		if err != nil {
			t.Fatalf("reading customer handler: %v", err)
		}
		if !strings.Contains(string(handlerBytes), "RegisterRoutes(mux *http.ServeMux)") {
			t.Errorf("expected *http.ServeMux in generated handler:\n%s", string(handlerBytes))
		}

		// Verify route discovery
		routesOut, routesErr := executeMakeCommand(root, "routes")
		if routesErr != nil {
			t.Fatalf("routes command failed: %v, out: %s", routesErr, routesOut)
		}
		if !strings.Contains(routesOut, "GET") {
			t.Errorf("expected routes in routes output: %s", routesOut)
		}
	})

	t.Run("gin handler generation", func(t *testing.T) {
		tempDir := t.TempDir()
		origDir, _ := os.Getwd()
		_ = os.Chdir(tempDir)
		t.Cleanup(func() { _ = os.Chdir(origDir) })

		_ = os.WriteFile("go.mod", []byte("module github.com/example/gin-app\n\ngo 1.22\n"), 0644)
		_ = os.WriteFile("loy.yaml", []byte("version: 1\nproject:\n  name: gin-app\ndefaults:\n  http: gin\n"), 0644)

		root := newCRUDTestRootCmd()
		out, err := executeMakeCommand(root, "make", "crud", "invoice", "number:string")
		if err != nil {
			t.Fatalf("make crud in gin project failed: %v, out: %s", err, out)
		}

		handlerBytes, err := os.ReadFile(filepath.Join(tempDir, "internal/invoice/transport/http/handler.go"))
		if err != nil {
			t.Fatalf("reading invoice handler: %v", err)
		}
		if !strings.Contains(string(handlerBytes), "RegisterRoutes(router *gin.RouterGroup)") {
			t.Errorf("expected *gin.RouterGroup in generated handler:\n%s", string(handlerBytes))
		}
		if !strings.Contains(string(handlerBytes), "c.Param(\"id\")") {
			t.Errorf("expected c.Param in generated handler:\n%s", string(handlerBytes))
		}

		// Verify route discovery
		routesOut, routesErr := executeMakeCommand(root, "routes")
		if routesErr != nil {
			t.Fatalf("routes command failed: %v, out: %s", routesErr, routesOut)
		}
		if !strings.Contains(routesOut, "GET") {
			t.Errorf("expected routes in routes output: %s", routesOut)
		}
	})
}
