package builtin

import (
	"context"
	"fmt"
	"time"

	"github.com/loy-go/loy/internal/generator"
	"github.com/loy-go/loy/internal/generator/model"
)

// IdempotencyGenerator scaffolds database migration and middleware for HTTP idempotency keys.
type IdempotencyGenerator struct {
	modulePath string
}

// NewIdempotencyGenerator constructs IdempotencyGenerator.
func NewIdempotencyGenerator(modulePath string) *IdempotencyGenerator {
	return &IdempotencyGenerator{modulePath: modulePath}
}

func (g *IdempotencyGenerator) Name() string {
	return "idempotency"
}

func (g *IdempotencyGenerator) Generate(ctx context.Context, input generator.Input) ([]model.Artifact, error) {
	data := NewBaseData("idempotency", g.modulePath, nil)
	renderer := GetRenderer()

	// 1. Migration
	migTmpl, err := ReadTemplateContext(ctx, "idempotency_migration.sql.tmpl")
	if err != nil {
		return nil, fmt.Errorf("reading idempotency migration template: %w", err)
	}
	migRendered, err := renderer.Render(ctx, "idempotency_migration.sql.tmpl", migTmpl, data)
	if err != nil {
		return nil, fmt.Errorf("rendering idempotency migration: %w", err)
	}
	timestamp := time.Now().UTC().Format("20060102150405")
	migPath := fmt.Sprintf("migrations/%s_create_idempotency_keys_table.sql", timestamp)

	// 2. Middleware
	mwTmpl, err := ReadTemplateContext(ctx, "idempotency_middleware.go.tmpl")
	if err != nil {
		return nil, fmt.Errorf("reading idempotency middleware template: %w", err)
	}
	mwRendered, err := renderer.RenderGo(ctx, "idempotency_middleware.go.tmpl", mwTmpl, data)
	if err != nil {
		return nil, fmt.Errorf("rendering idempotency middleware: %w", err)
	}

	return []model.Artifact{
		{
			Path:        migPath,
			Content:     migRendered,
			Ownership:   model.DeveloperOwned,
			Permissions: 0644,
		},
		{
			Path:        "internal/platform/middleware/idempotency.go",
			Content:     mwRendered,
			Ownership:   model.DeveloperOwned,
			Permissions: 0644,
		},
	}, nil
}
