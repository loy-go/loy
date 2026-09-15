package builtin

import (
	"context"
	"fmt"

	"github.com/loy-go/loy/internal/generator"
	"github.com/loy-go/loy/internal/generator/model"
)

// TenantGenerator scaffolds PostgreSQL RLS multi-tenancy context and migrations.
type TenantGenerator struct {
	modulePath   string
	withDatabase bool
}

// NewTenantGenerator constructs TenantGenerator.
func NewTenantGenerator(modulePath string) *TenantGenerator {
	return &TenantGenerator{
		modulePath:   modulePath,
		withDatabase: true,
	}
}

// WithDatabase sets whether database migrations should be scaffolded.
func (g *TenantGenerator) WithDatabase(withDB bool) *TenantGenerator {
	g.withDatabase = withDB
	return g
}

func (g *TenantGenerator) Name() string {
	return "tenant"
}

func (g *TenantGenerator) Generate(ctx context.Context, input generator.Input) ([]model.Artifact, error) {
	data := NewBaseData("tenant", g.modulePath, nil)
	renderer := GetRenderer()

	// 1. Context and middleware
	ctxTmpl, err := ReadTemplate("tenant_context.go.tmpl")
	if err != nil {
		return nil, fmt.Errorf("reading tenant_context template: %w", err)
	}
	ctxRendered, err := renderer.RenderGo(ctx, "tenant_context.go.tmpl", ctxTmpl, data)
	if err != nil {
		return nil, fmt.Errorf("rendering tenant_context: %w", err)
	}

	artifacts := []model.Artifact{
		{
			Path:        "internal/platform/tenant/context.go",
			Content:     ctxRendered,
			Ownership:   model.DeveloperOwned,
			Permissions: 0644,
		},
	}

	// 2. Initial tenancy migration
	if g.withDatabase {
		migTmpl, err := ReadTemplate("tenant_migration.sql.tmpl")
		if err != nil {
			return nil, fmt.Errorf("reading tenant_migration template: %w", err)
		}
		migRendered, err := renderer.Render(ctx, "tenant_migration.sql.tmpl", migTmpl, data)
		if err != nil {
			return nil, fmt.Errorf("rendering tenant_migration: %w", err)
		}
		artifacts = append(artifacts, model.Artifact{
			Path:        "migrations/00001_init_tenancy.sql",
			Content:     migRendered,
			Ownership:   model.DeveloperOwned,
			Permissions: 0644,
		})
	}

	return artifacts, nil
}
