package builtin_test

import (
	"context"
	"strings"
	"testing"

	"github.com/loy-go/loy/internal/generator"
	"github.com/loy-go/loy/internal/generator/builtin"
)

func TestAtomicGenerators(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	mod := "github.com/example/myapp"

	t.Run("model generator", func(t *testing.T) {
		gen := builtin.NewModelGenerator(mod)
		if gen.Name() != "model" {
			t.Fatalf("expected name model, got %s", gen.Name())
		}

		artifacts, err := gen.Generate(ctx, generator.Input{
			Name: "user",
			Args: map[string]string{
				"fields": "email:string:unique age:int:optional",
			},
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(artifacts) != 1 {
			t.Fatalf("expected 1 artifact, got %d", len(artifacts))
		}
		if artifacts[0].Path != "internal/user/domain/user.go" {
			t.Fatalf("expected path internal/user/domain/user.go, got %s", artifacts[0].Path)
		}
		content := string(artifacts[0].Content)
		t.Logf("Generated model content:\n%s", content)
		if !strings.Contains(content, "type User struct") {
			t.Fatalf("missing User struct")
		}
		if !strings.Contains(content, "Email     string") {
			t.Fatalf("missing Email field")
		}
		if !strings.Contains(content, "Age       *int") {
			t.Fatalf("missing Age *int field")
		}
	})

	t.Run("repository generator", func(t *testing.T) {
		gen := builtin.NewRepositoryGenerator(mod)
		if gen.Name() != "repository" {
			t.Fatalf("expected name repository, got %s", gen.Name())
		}

		artifacts, err := gen.Generate(ctx, generator.Input{Name: "user"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(artifacts) != 2 {
			t.Fatalf("expected 2 artifacts, got %d", len(artifacts))
		}
		if artifacts[0].Path != "internal/user/domain/repository.go" {
			t.Fatalf("expected path internal/user/domain/repository.go, got %s", artifacts[0].Path)
		}
		if artifacts[1].Path != "internal/user/repository/pg_adapter.go" {
			t.Fatalf("expected path internal/user/repository/pg_adapter.go, got %s", artifacts[1].Path)
		}

		ifaceContent := string(artifacts[0].Content)
		if !strings.Contains(ifaceContent, "type Repository interface") {
			t.Fatalf("missing Repository interface")
		}

		adapterContent := string(artifacts[1].Content)
		if !strings.Contains(adapterContent, "type PostgresRepository struct") {
			t.Fatalf("missing PostgresRepository struct")
		}
	})

	t.Run("service generator", func(t *testing.T) {
		gen := builtin.NewServiceGenerator(mod)
		artifacts, err := gen.Generate(ctx, generator.Input{Name: "user"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(artifacts) != 1 {
			t.Fatalf("expected 1 artifact, got %d", len(artifacts))
		}
		content := string(artifacts[0].Content)
		if !strings.Contains(content, "type Service struct") {
			t.Fatalf("missing Service struct")
		}
	})

	t.Run("handler generator", func(t *testing.T) {
		gen := builtin.NewHandlerGenerator(mod)
		artifacts, err := gen.Generate(ctx, generator.Input{Name: "user"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(artifacts) != 1 {
			t.Fatalf("expected 1 artifact, got %d", len(artifacts))
		}
		content := string(artifacts[0].Content)
		if !strings.Contains(content, "RegisterRoutes(router fiber.Router)") {
			t.Fatalf("missing RegisterRoutes")
		}
	})

	t.Run("handler generator chi", func(t *testing.T) {
		gen := builtin.NewHandlerGenerator(mod).WithHTTPFramework("chi")
		artifacts, err := gen.Generate(ctx, generator.Input{Name: "user"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(artifacts) != 1 {
			t.Fatalf("expected 1 artifact, got %d", len(artifacts))
		}
		content := string(artifacts[0].Content)
		if !strings.Contains(content, "RegisterRoutes(router chi.Router)") {
			t.Fatalf("missing Chi RegisterRoutes in %s", content)
		}
		if !strings.Contains(content, "chi.URLParam(r, \"id\")") {
			t.Fatalf("missing chi.URLParam in %s", content)
		}
	})

	t.Run("handler generator nethttp", func(t *testing.T) {
		gen := builtin.NewHandlerGenerator(mod).WithHTTPFramework("nethttp")
		artifacts, err := gen.Generate(ctx, generator.Input{Name: "user"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(artifacts) != 1 {
			t.Fatalf("expected 1 artifact, got %d", len(artifacts))
		}
		content := string(artifacts[0].Content)
		if !strings.Contains(content, "RegisterRoutes(mux *http.ServeMux)") {
			t.Fatalf("missing nethttp RegisterRoutes in %s", content)
		}
		if !strings.Contains(content, "r.PathValue(\"id\")") {
			t.Fatalf("missing r.PathValue in %s", content)
		}
	})

	t.Run("handler generator gin", func(t *testing.T) {
		gen := builtin.NewHandlerGenerator(mod).WithHTTPFramework("gin")
		artifacts, err := gen.Generate(ctx, generator.Input{Name: "user"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(artifacts) != 1 {
			t.Fatalf("expected 1 artifact, got %d", len(artifacts))
		}
		content := string(artifacts[0].Content)
		if !strings.Contains(content, "RegisterRoutes(router *gin.RouterGroup)") {
			t.Fatalf("missing gin RegisterRoutes in %s", content)
		}
		if !strings.Contains(content, "c.Param(\"id\")") {
			t.Fatalf("missing c.Param in %s", content)
		}
	})

	t.Run("request generator", func(t *testing.T) {
		gen := builtin.NewRequestGenerator(mod)
		artifacts, err := gen.Generate(ctx, generator.Input{
			Name: "user",
			Args: map[string]string{
				"fields": "title:string:required",
			},
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(artifacts) != 1 {
			t.Fatalf("expected 1 artifact, got %d", len(artifacts))
		}
		content := string(artifacts[0].Content)
		if !strings.Contains(content, "type UserRequest struct") {
			t.Fatalf("missing UserRequest struct")
		}
	})

	t.Run("resource generator", func(t *testing.T) {
		gen := builtin.NewResourceGenerator(mod)
		artifacts, err := gen.Generate(ctx, generator.Input{
			Name: "user",
			Args: map[string]string{
				"fields": "username:string",
			},
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(artifacts) != 1 {
			t.Fatalf("expected 1 artifact, got %d", len(artifacts))
		}
		content := string(artifacts[0].Content)
		if !strings.Contains(content, "type UserResource struct") {
			t.Fatalf("missing UserResource struct")
		}
	})

	t.Run("job generator", func(t *testing.T) {
		gen := builtin.NewJobGenerator(mod)
		artifacts, err := gen.Generate(ctx, generator.Input{Name: "user_sync"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(artifacts) != 3 {
			t.Fatalf("expected 3 artifacts (job file + 2 wiring regions), got %d", len(artifacts))
		}
		content := string(artifacts[0].Content)
		if !strings.Contains(content, "TypeUserSyncProcess") {
			t.Fatalf("missing task type")
		}
	})

	t.Run("event & listener generator", func(t *testing.T) {
		eventGen := builtin.NewEventGenerator(mod)
		listenerGen := builtin.NewListenerGenerator(mod)

		evArts, err := eventGen.Generate(ctx, generator.Input{Name: "order"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(evArts) != 1 || !strings.Contains(string(evArts[0].Content), "type OrderCreatedEvent struct") {
			t.Fatalf("unexpected event artifact: %+v", evArts)
		}

		lisArts, err := listenerGen.Generate(ctx, generator.Input{Name: "order"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(lisArts) != 1 || !strings.Contains(string(lisArts[0].Content), "type OrderCreatedListener struct") {
			t.Fatalf("unexpected listener artifact: %+v", lisArts)
		}
	})

	t.Run("policy generator", func(t *testing.T) {
		gen := builtin.NewPolicyGenerator(mod)
		arts, err := gen.Generate(ctx, generator.Input{Name: "invoice"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(arts) != 1 || !strings.Contains(string(arts[0].Content), "type InvoicePolicy struct") {
			t.Fatalf("unexpected policy artifact: %+v", arts)
		}
	})

	t.Run("test generator", func(t *testing.T) {
		gen := builtin.NewTestGenerator(mod)
		arts, err := gen.Generate(ctx, generator.Input{Name: "invoice"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(arts) != 1 || !strings.Contains(string(arts[0].Content), "TestInvoiceService_GetByID") {
			t.Fatalf("unexpected test artifact: %+v", arts)
		}
	})

	t.Run("auth generator", func(t *testing.T) {
		gen := builtin.NewAuthGenerator(mod)
		arts, err := gen.Generate(ctx, generator.Input{Name: "auth"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(arts) != 4 {
			t.Fatalf("expected 4 auth artifacts, got %d", len(arts))
		}
	})

	t.Run("tenant generator", func(t *testing.T) {
		gen := builtin.NewTenantGenerator(mod).WithDatabase(true)
		arts, err := gen.Generate(ctx, generator.Input{Name: "tenant"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(arts) != 2 {
			t.Fatalf("expected 2 tenant artifacts, got %d", len(arts))
		}
		if arts[0].Path != "internal/platform/tenant/context.go" {
			t.Errorf("expected tenant context.go, got: %s", arts[0].Path)
		}
		if arts[1].Path != "migrations/00001_init_tenancy.sql" {
			t.Errorf("expected tenancy migration, got: %s", arts[1].Path)
		}
	})

	t.Run("metrics generator", func(t *testing.T) {
		gen := builtin.NewMetricsGenerator(mod)
		arts, err := gen.Generate(ctx, generator.Input{Name: "app"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(arts) != 3 {
			t.Fatalf("expected 3 metrics artifacts, got %d", len(arts))
		}
		if arts[0].Path != "internal/platform/metrics/metrics.go" {
			t.Errorf("expected metrics.go, got: %s", arts[0].Path)
		}
		if arts[1].Path != "deploy/grafana/dashboard.json" {
			t.Errorf("expected dashboard.json, got: %s", arts[1].Path)
		}
		if arts[2].Path != "deploy/prometheus/prometheus.yml" {
			t.Errorf("expected prometheus.yml, got: %s", arts[2].Path)
		}
	})

	t.Run("grpc generator", func(t *testing.T) {
		gen := builtin.NewGRPCGenerator(mod)
		arts, err := gen.Generate(ctx, generator.Input{Name: "candidate"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(arts) != 2 {
			t.Fatalf("expected 2 grpc artifacts, got %d", len(arts))
		}
	})

	t.Run("ws generator", func(t *testing.T) {
		gen := builtin.NewWebSocketGenerator(mod)
		arts, err := gen.Generate(ctx, generator.Input{Name: "interview"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(arts) != 3 {
			t.Fatalf("expected 3 ws artifacts, got %d", len(arts))
		}
	})

	t.Run("outbox generator", func(t *testing.T) {
		gen := builtin.NewOutboxGenerator(mod)
		arts, err := gen.Generate(ctx, generator.Input{Name: "outbox"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(arts) != 3 {
			t.Fatalf("expected 3 outbox artifacts, got %d", len(arts))
		}
	})

	t.Run("seeder generator", func(t *testing.T) {
		gen := builtin.NewSeederGenerator(mod)
		arts, err := gen.Generate(ctx, generator.Input{Name: "candidate"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(arts) != 1 {
			t.Fatalf("expected 1 seeder artifact, got %d", len(arts))
		}
	})
}

func TestCompositeGenerators(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	mod := "github.com/example/myapp"

	t.Run("feature generator", func(t *testing.T) {
		gen := builtin.NewFeatureGenerator(mod)
		if gen.Name() != "feature" {
			t.Fatalf("expected name feature, got %s", gen.Name())
		}

		artifacts, err := gen.Generate(ctx, generator.Input{
			Name: "order",
			Args: map[string]string{
				"fields": "total:float:required status:string",
			},
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(artifacts) != 11 {
			t.Fatalf("expected 11 artifacts, got %d", len(artifacts))
		}
	})

	t.Run("crud generator", func(t *testing.T) {
		gen := builtin.NewCRUDGenerator(mod)
		if gen.Name() != "crud" {
			t.Fatalf("expected name crud, got %s", gen.Name())
		}

		artifacts, err := gen.Generate(ctx, generator.Input{
			Name: "product",
			Args: map[string]string{
				"fields": "name:string:unique price:float64",
			},
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(artifacts) != 13 {
			t.Fatalf("expected 13 artifacts, got %d", len(artifacts))
		}
		if !strings.Contains(artifacts[0].Path, "migrations/") {
			t.Fatalf("expected migration path, got %s", artifacts[0].Path)
		}
		if !strings.Contains(artifacts[1].Path, "queries/products.sql") {
			t.Fatalf("expected queries path, got %s", artifacts[1].Path)
		}
	})
}
