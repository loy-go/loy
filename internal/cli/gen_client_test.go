package cli_test

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/loy-go/loy/internal/cli"
)

func TestGenClientCommand(t *testing.T) {
	tempDir := t.TempDir()

	// Create sample project structure
	_ = os.MkdirAll(filepath.Join(tempDir, "internal/user/transport/http"), 0755)
	dtoGo := `package http

type UserRequest struct {
	Name  string ` + "`json:\"name\"`" + `
	Email string ` + "`json:\"email\"`" + `
}

type UserResource struct {
	ID    int64  ` + "`json:\"id\"`" + `
	Name  string ` + "`json:\"name\"`" + `
	Email string ` + "`json:\"email\"`" + `
}
`
	_ = os.WriteFile(filepath.Join(tempDir, "internal/user/transport/http/dto.go"), []byte(dtoGo), 0644)

	routesGo := `package http

import "net/http"

func RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /users", nil)
	mux.HandleFunc("POST /users", nil)
}
`
	_ = os.WriteFile(filepath.Join(tempDir, "internal/user/transport/http/routes.go"), []byte(routesGo), 0644)

	t.Run("default text output", func(t *testing.T) {
		outDir := filepath.Join(tempDir, "frontend/src/api")
		cmd := cli.NewRootCmd()
		var stdout bytes.Buffer
		cmd.SetOut(&stdout)
		cmd.SetArgs([]string{"gen", "client", tempDir, "--out", outDir})

		if err := cmd.Execute(); err != nil {
			t.Fatalf("loy gen client failed: %v", err)
		}

		outStr := stdout.String()
		if !strings.Contains(outStr, "Successfully generated TypeScript client") {
			t.Errorf("expected success message, got: %s", outStr)
		}

		if _, err := os.Stat(filepath.Join(outDir, "types.ts")); os.IsNotExist(err) {
			t.Errorf("expected types.ts to exist in %s", outDir)
		}
		if _, err := os.Stat(filepath.Join(outDir, "client.ts")); os.IsNotExist(err) {
			t.Errorf("expected client.ts to exist in %s", outDir)
		}
	})

	t.Run("json output", func(t *testing.T) {
		outDir := filepath.Join(tempDir, "client_json")
		cmd := cli.NewRootCmd()
		var stdout bytes.Buffer
		cmd.SetOut(&stdout)
		cmd.SetArgs([]string{"generate", "client", tempDir, "--out", outDir, "--json"})

		if err := cmd.Execute(); err != nil {
			t.Fatalf("loy generate client --json failed: %v", err)
		}

		var res map[string]string
		if err := json.Unmarshal(stdout.Bytes(), &res); err != nil {
			t.Fatalf("failed to parse json output: %v, raw: %s", err, stdout.String())
		}

		if filepath.Clean(res["types_path"]) != filepath.Clean(filepath.Join(outDir, "types.ts")) {
			t.Errorf("expected types_path, got %s", res["types_path"])
		}
		if filepath.Clean(res["client_path"]) != filepath.Clean(filepath.Join(outDir, "client.ts")) {
			t.Errorf("expected client_path, got %s", res["client_path"])
		}
	})
}
