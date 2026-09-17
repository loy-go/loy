package typescript

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/loy-go/loy/internal/filesystem"
	"github.com/loy-go/loy/internal/route"
)

func TestParseDTOs_And_GenerateTypes(t *testing.T) {
	fs := filesystem.NewMemFileSystem()

	goSource := `package request

import (
	"time"
	"github.com/google/uuid"
)

type CreateUserRequest struct {
	Name     string             ` + "`json:\"name\"`" + `
	Email    string             ` + "`json:\"email\"`" + `
	Age      *int               ` + "`json:\"age,omitempty\"`" + `
	IsActive bool               ` + "`json:\"is_active\"`" + `
	Tags     []string           ` + "`json:\"tags\"`" + `
	Metadata map[string]string  ` + "`json:\"metadata\"`" + `
	Ignored  string             ` + "`json:\"-\"`" + `
	unexported string
}

type UserResource struct {
	ID        int64     ` + "`json:\"id\"`" + `
	PublicID  uuid.UUID ` + "`json:\"public_id\"`" + `
	Name      string    ` + "`json:\"name\"`" + `
	Email     string    ` + "`json:\"email\"`" + `
	CreatedAt time.Time ` + "`json:\"created_at\"`" + `
}
`
	if err := fs.MkdirAll("internal/user/transport", 0755); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}
	if err := fs.WriteFile("internal/user/transport/user_dto.go", []byte(goSource), 0644); err != nil {
		t.Fatalf("write failed: %v", err)
	}

	types, err := ParseDTOs(fs, ".")
	if err != nil {
		t.Fatalf("ParseDTOs failed: %v", err)
	}

	if len(types) != 2 {
		t.Fatalf("expected 2 types, got %d", len(types))
	}

	typesTS := GenerateTypes(types)
	if !strings.Contains(typesTS, "export interface CreateUserRequest {") {
		t.Errorf("expected CreateUserRequest interface, got: %s", typesTS)
	}
	if !strings.Contains(typesTS, "  name: string;") {
		t.Errorf("expected name: string, got: %s", typesTS)
	}
	if !strings.Contains(typesTS, "  age?: number | null;") {
		t.Errorf("expected age?: number | null, got: %s", typesTS)
	}
	if !strings.Contains(typesTS, "  tags: string[];") {
		t.Errorf("expected tags: string[], got: %s", typesTS)
	}
	if !strings.Contains(typesTS, "  metadata: Record<string, string>;") {
		t.Errorf("expected metadata: Record<string, string>, got: %s", typesTS)
	}
	if strings.Contains(typesTS, "ignored") || strings.Contains(typesTS, "unexported") {
		t.Errorf("should not contain ignored or unexported fields, got: %s", typesTS)
	}

	if !strings.Contains(typesTS, "export interface UserResource {") {
		t.Errorf("expected UserResource interface, got: %s", typesTS)
	}
	if !strings.Contains(typesTS, "  public_id: string;") {
		t.Errorf("expected public_id: string, got: %s", typesTS)
	}
	if !strings.Contains(typesTS, "  created_at: string;") {
		t.Errorf("expected created_at: string, got: %s", typesTS)
	}
}

func TestGenerateClient(t *testing.T) {
	routes := []route.RouteInfo{
		{Method: "GET", Path: "/api/v1/users", Handler: "h.List"},
		{Method: "GET", Path: "/api/v1/users/:id", Handler: "h.GetByID"},
		{Method: "POST", Path: "/api/v1/users", Handler: "h.Create"},
		{Method: "PUT", Path: "/api/v1/users/:id", Handler: "h.Update"},
		{Method: "DELETE", Path: "/api/v1/users/:id", Handler: "h.Delete"},
		{Method: "POST", Path: "/posts/{postId}/comments", Handler: "h.AddComment"},
	}

	types := []TypeDefinition{
		{
			Name: "UserRequest",
			Fields: []FieldDefinition{
				{Name: "name", Type: "string"},
			},
		},
		{
			Name: "UserResource",
			Fields: []FieldDefinition{
				{Name: "id", Type: "number"},
				{Name: "name", Type: "string"},
			},
		},
	}

	clientTS := GenerateClient(routes, types)

	if !strings.Contains(clientTS, "import type { UserRequest, UserResource } from './types';") {
		t.Errorf("expected imports from types, got: %s", clientTS)
	}

	if !strings.Contains(clientTS, "export class LoyClient {") {
		t.Errorf("expected class LoyClient, got: %s", clientTS)
	}

	// Verify methods
	if !strings.Contains(clientTS, "async listUsers(params?: { limit?: number; offset?: number; [key: string]: any }): Promise<UserResource[]>") {
		t.Errorf("expected listUsers method, got: %s", clientTS)
	}
	if !strings.Contains(clientTS, "async getUser(id: string | number): Promise<UserResource>") {
		t.Errorf("expected getUser method, got: %s", clientTS)
	}
	if !strings.Contains(clientTS, "async createUser(data: UserRequest): Promise<UserResource>") {
		t.Errorf("expected createUser method, got: %s", clientTS)
	}
	if !strings.Contains(clientTS, "async updateUser(id: string | number, data: UserRequest): Promise<UserResource>") {
		t.Errorf("expected updateUser method, got: %s", clientTS)
	}
	if !strings.Contains(clientTS, "async deleteUser(id: string | number): Promise<void>") {
		t.Errorf("expected deleteUser method, got: %s", clientTS)
	}
	if !strings.Contains(clientTS, "postId: string | number") {
		t.Errorf("expected postId parameter, got: %s", clientTS)
	}
}

func TestGenerate_EndToEnd(t *testing.T) {
	fs := filesystem.NewMemFileSystem()

	// Write sample request and resource
	_ = fs.MkdirAll("internal/item/transport/http", 0755)
	dtoContent := `package http

type ItemRequest struct {
	Title string ` + "`json:\"title\"`" + `
}

type ItemResource struct {
	ID    int64  ` + "`json:\"id\"`" + `
	Title string ` + "`json:\"title\"`" + `
}
`
	_ = fs.WriteFile("internal/item/transport/http/dto.go", []byte(dtoContent), 0644)

	handlerContent := `package http

import "net/http"

func RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /items", nil)
	mux.HandleFunc("POST /items", nil)
	mux.HandleFunc("GET /items/{id}", nil)
}
`
	_ = fs.WriteFile("internal/item/transport/http/routes.go", []byte(handlerContent), 0644)

	result, err := Generate(fs, GenerateOptions{
		RootDir: ".",
		OutDir:  "client",
	})
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	if filepath.Clean(result.TypesPath) != filepath.Clean(filepath.Join("client", "types.ts")) {
		t.Errorf("expected types path client/types.ts, got %s", result.TypesPath)
	}
	if filepath.Clean(result.ClientPath) != filepath.Clean(filepath.Join("client", "client.ts")) {
		t.Errorf("expected client path client/client.ts, got %s", result.ClientPath)
	}

	typesData, err := fs.ReadFile(result.TypesPath)
	if err != nil || !strings.Contains(string(typesData), "ItemRequest") {
		t.Errorf("expected types.ts to contain ItemRequest, got: %s", string(typesData))
	}

	clientData, err := fs.ReadFile(result.ClientPath)
	if err != nil || !strings.Contains(string(clientData), "class LoyClient") {
		t.Errorf("expected client.ts to contain LoyClient, got: %s", string(clientData))
	}
}

func TestEmptyTypesAndRoutes(t *testing.T) {
	typesTS := GenerateTypes(nil)
	if !strings.Contains(typesTS, "export {};") {
		t.Errorf("expected export {}; for empty types, got: %s", typesTS)
	}

	clientTS := GenerateClient(nil, nil)
	if !strings.Contains(clientTS, "class LoyClient") {
		t.Errorf("expected class LoyClient even for empty routes, got: %s", clientTS)
	}
}

func TestGenerate_TypeScriptCompilerVerification(t *testing.T) {
	tempDir := t.TempDir()
	osFS := filesystem.NewOSFileSystem()

	// Write mock Go files
	_ = osFS.MkdirAll(filepath.Join(tempDir, "internal", "user", "transport"), 0755)
	dtoGo := `package transport
type UserRequest struct {
	Name string ` + "`json:\"name\"`" + `
	Email string ` + "`json:\"email\"`" + `
}
type UserResource struct {
	ID int64 ` + "`json:\"id\"`" + `
	Name string ` + "`json:\"name\"`" + `
	Email string ` + "`json:\"email\"`" + `
}
`
	_ = osFS.WriteFile(filepath.Join(tempDir, "internal", "user", "transport", "user.go"), []byte(dtoGo), 0644)

	routesGo := `package transport
import "net/http"
func Routes(mux *http.ServeMux) {
	mux.HandleFunc("GET /users", nil)
	mux.HandleFunc("POST /users", nil)
	mux.HandleFunc("GET /users/{id}", nil)
	mux.HandleFunc("DELETE /users/{id}", nil)
}
`
	_ = osFS.WriteFile(filepath.Join(tempDir, "internal", "user", "transport", "routes.go"), []byte(routesGo), 0644)

	clientOut := filepath.Join(tempDir, "client")
	result, err := Generate(osFS, GenerateOptions{
		RootDir: tempDir,
		OutDir:  clientOut,
	})
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	if result.TypesSource == "" || result.ClientSource == "" {
		t.Fatalf("expected non-empty sources")
	}
}

func TestGenerate_PathTraversalEscape(t *testing.T) {
	fs := filesystem.NewMemFileSystem()
	_ = fs.MkdirAll("/workspace", 0755)

	_, err := Generate(fs, GenerateOptions{
		RootDir: "/workspace",
		OutDir:  "../../etc",
	})
	if err == nil {
		t.Errorf("expected error on path traversal in OutDir, got nil")
	}
}

func TestGenerate_EndpointStartingWithV(t *testing.T) {
	routes := []route.RouteInfo{
		{Method: "GET", Path: "/api/v1/videos", Handler: "h.List"},
		{Method: "GET", Path: "/api/v1/videos/:id", Handler: "h.GetByID"},
	}
	clientTS := GenerateClient(routes, nil)
	if !strings.Contains(clientTS, "listVideos(") {
		t.Errorf("expected listVideos method for /videos endpoint, got:\n%s", clientTS)
	}
	if strings.Contains(clientTS, "listRoots(") {
		t.Errorf("should not generate listRoots for /videos, got:\n%s", clientTS)
	}
}

func TestGenerate_NullableWithoutOmitempty(t *testing.T) {
	fs := filesystem.NewMemFileSystem()
	dtoContent := `package transport

type ScoreRecord struct {
	Score *int ` + "`json:\"score\"`" + `
}
`
	_ = fs.MkdirAll("internal/game", 0755)
	_ = fs.WriteFile("internal/game/score.go", []byte(dtoContent), 0644)

	types, err := ParseDTOs(fs, ".")
	if err != nil {
		t.Fatalf("ParseDTOs failed: %v", err)
	}
	typesTS := GenerateTypes(types)
	if !strings.Contains(typesTS, "  score: number | null;") {
		t.Errorf("expected 'score: number | null;', got:\n%s", typesTS)
	}
}
