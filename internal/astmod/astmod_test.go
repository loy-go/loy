package astmod

import (
	"fmt"
	"go/parser"
	"go/token"
	"math/rand"
	"strings"
	"testing"
)

func TestInsertField(t *testing.T) {
	initialSrc := `package model

// User represents a system user.
type User struct {
	// ID is the unique database identifier.
	ID int64 ` + "`json:\"id\"`" + `
	// Name is the display name.
	Name string ` + "`json:\"name\"`" + `
}
`

	// 1. Insert a simple field
	modified, err := InsertField([]byte(initialSrc), "User", "Email", "string", `json:"email"`)
	if err != nil {
		t.Fatalf("InsertField failed: %v", err)
	}

	res := string(modified)
	if !strings.Contains(res, "Email string `json:\"email\"`") {
		t.Errorf("expected Email field in result:\n%s", res)
	}
	// Verify comments are retained
	if !strings.Contains(res, "// User represents a system user.") || !strings.Contains(res, "// ID is the unique database identifier.") {
		t.Errorf("expected comments preserved:\n%s", res)
	}

	// 2. Insert pointer and slice types
	modified2, err := InsertField(modified, "User", "Roles", "[]string", `json:"roles"`)
	if err != nil {
		t.Fatalf("InsertField with slice failed: %v", err)
	}
	if !strings.Contains(string(modified2), "Roles []string `json:\"roles\"`") {
		t.Errorf("expected Roles slice in result:\n%s", string(modified2))
	}

	// 3. Duplicate field error
	_, err = InsertField(modified2, "User", "Email", "string", "")
	if err == nil {
		t.Errorf("expected error for duplicate field, got nil")
	}

	// 4. Missing struct error
	_, err = InsertField(modified2, "NonExistent", "Foo", "string", "")
	if err == nil {
		t.Errorf("expected error for missing struct, got nil")
	}

	// 5. Invalid type error
	_, err = InsertField(modified2, "User", "Bad", "invalid syntax {{", "")
	if err == nil {
		t.Errorf("expected error for invalid type, got nil")
	}
}

func TestAddRoute(t *testing.T) {
	initialSrc := `package http

import "github.com/gofiber/fiber/v2"

type Handler struct{}

// RegisterRoutes registers endpoints on router.
func (h *Handler) RegisterRoutes(router fiber.Router) {
	router.Get("/users", h.List)
}
`
	modified, err := AddRoute([]byte(initialSrc), "POST", "/users", "h.Create")
	if err != nil {
		t.Fatalf("AddRoute failed: %v", err)
	}

	res := string(modified)
	if !strings.Contains(res, `router.Post("/users", h.Create)`) {
		t.Errorf("expected router.Post in result:\n%s", res)
	}
	if !strings.Contains(res, "// RegisterRoutes registers endpoints on router.") {
		t.Errorf("expected comments preserved in result:\n%s", res)
	}

	// Test net/http mux
	muxSrc := `package http

import "net/http"

func SetupRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /items", nil)
}
`
	muxModified, err := AddRoute([]byte(muxSrc), "DELETE", "/items/{id}", "nil")
	if err != nil {
		t.Fatalf("AddRoute for mux failed: %v", err)
	}
	if !strings.Contains(string(muxModified), `mux.HandleFunc("DELETE /items/{id}", nil)`) {
		t.Errorf("expected mux.HandleFunc DELETE in result:\n%s", string(muxModified))
	}
}

func TestBindDependency(t *testing.T) {
	initialWiring := `package app

func WireApp(a *App) error {
	// Initialize database
	db := a.db

	return nil
}
`
	modified, err := BindDependency([]byte(initialWiring), "userRepo.NewPostgresRepository(db)", "userRepo")
	if err != nil {
		t.Fatalf("BindDependency failed: %v", err)
	}

	res := string(modified)
	if !strings.Contains(res, "userRepo, _ := userRepo.NewPostgresRepository(db)") {
		t.Errorf("expected constructor call in result:\n%s", res)
	}

	// Idempotency: binding again should return unchanged
	modifiedAgain, err := BindDependency(modified, "userRepo.NewPostgresRepository(db)", "userRepo")
	if err != nil {
		t.Fatalf("idempotent BindDependency failed: %v", err)
	}
	if string(modifiedAgain) != string(modified) {
		t.Errorf("expected idempotent BindDependency")
	}
}

func TestFuzzASTMod_50Mutations(t *testing.T) {
	src := `package testpkg

// HeavyCommentStruct is documented extensively.
// Line 2 of documentation.
type HeavyCommentStruct struct {
	// BaseField initial field
	BaseField string ` + "`json:\"base_field\"`" + `
}

// RegisterRoutes handles route bindings.
func RegisterRoutes(router interface{ Get(string, any); Post(string, any) }) {
	router.Get("/base", nil)
}
`

	curr := []byte(src)
	rnd := rand.New(rand.NewSource(42))

	types := []string{"string", "int", "int64", "bool", "float64", "[]string", "*int", "map[string]any"}

	// Perform 50 sequential field and route insertions
	for i := 0; i < 50; i++ {
		fieldName := fmt.Sprintf("Field%d", i)
		fieldType := types[rnd.Intn(len(types))]
		tag := fmt.Sprintf("json:\"field_%d,omitempty\"", i)

		var err error
		curr, err = InsertField(curr, "HeavyCommentStruct", fieldName, fieldType, tag)
		if err != nil {
			t.Fatalf("fuzz InsertField failed on iteration %d: %v", i, err)
		}

		if i%2 == 0 {
			routePath := fmt.Sprintf("/endpoint/%d", i)
			method := "GET"
			if i%4 == 0 {
				method = "POST"
			}
			curr, err = AddRoute(curr, method, routePath, "nil")
			if err != nil {
				t.Fatalf("fuzz AddRoute failed on iteration %d: %v", i, err)
			}
		}
	}

	// Verify the final Go code parses cleanly with standard Go parser
	fset := token.NewFileSet()
	_, parseErr := parser.ParseFile(fset, "test.go", curr, parser.ParseComments)
	if parseErr != nil {
		t.Fatalf("fuzz generated code failed standard go/parser check: %v\nCode:\n%s", parseErr, string(curr))
	}

	// Verify comments were never dropped
	finalStr := string(curr)
	if !strings.Contains(finalStr, "// HeavyCommentStruct is documented extensively.") {
		t.Errorf("header comment was dropped during fuzzing")
	}
	if !strings.Contains(finalStr, "// Line 2 of documentation.") {
		t.Errorf("second line comment was dropped during fuzzing")
	}
	if !strings.Contains(finalStr, "// BaseField initial field") {
		t.Errorf("field comment was dropped during fuzzing")
	}
	if !strings.Contains(finalStr, "// RegisterRoutes handles route bindings.") {
		t.Errorf("func comment was dropped during fuzzing")
	}
}
