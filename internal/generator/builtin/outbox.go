package builtin

import (
	"context"
	"fmt"
	"time"

	"github.com/loy-go/loy/internal/generator"
	"github.com/loy-go/loy/internal/generator/model"
)

// OutboxGenerator creates Transactional Outbox migration, store, and dispatcher.
type OutboxGenerator struct {
	modulePath string
}

// NewOutboxGenerator constructs OutboxGenerator.
func NewOutboxGenerator(modulePath string) *OutboxGenerator {
	return &OutboxGenerator{modulePath: modulePath}
}

func (g *OutboxGenerator) Name() string {
	return "outbox"
}

func (g *OutboxGenerator) Generate(ctx context.Context, input generator.Input) ([]model.Artifact, error) {
	data := NewBaseData("outbox", g.modulePath, nil)
	renderer := GetRenderer()

	// 1. Migration
	migTmpl, err := ReadTemplate("outbox_migration.sql.tmpl")
	if err != nil {
		return nil, fmt.Errorf("reading outbox_migration template: %w", err)
	}
	migRendered, err := renderer.Render(ctx, "outbox_migration.sql.tmpl", migTmpl, data)
	if err != nil {
		return nil, fmt.Errorf("rendering outbox migration: %w", err)
	}
	timestamp := time.Now().UTC().Format("20060102150405")
	migPath := fmt.Sprintf("migrations/%s_create_outbox_events_table.sql", timestamp)

	// 2. Store
	storeTmpl, err := ReadTemplate("outbox_repository.go.tmpl")
	if err != nil {
		return nil, fmt.Errorf("reading outbox_repository template: %w", err)
	}
	storeRendered, err := renderer.RenderGo(ctx, "outbox_repository.go.tmpl", storeTmpl, data)
	if err != nil {
		return nil, fmt.Errorf("rendering outbox store: %w", err)
	}

	// 3. Dispatcher
	dispTmpl, err := ReadTemplate("outbox_dispatcher.go.tmpl")
	if err != nil {
		return nil, fmt.Errorf("reading outbox_dispatcher template: %w", err)
	}
	dispRendered, err := renderer.RenderGo(ctx, "outbox_dispatcher.go.tmpl", dispTmpl, data)
	if err != nil {
		return nil, fmt.Errorf("rendering outbox dispatcher: %w", err)
	}

	return []model.Artifact{
		{
			Path:        migPath,
			Content:     migRendered,
			Ownership:   model.DeveloperOwned,
			Permissions: 0644,
		},
		{
			Path:        "internal/platform/outbox/store.go",
			Content:     storeRendered,
			Ownership:   model.DeveloperOwned,
			Permissions: 0644,
		},
		{
			Path:        "internal/platform/outbox/dispatcher.go",
			Content:     dispRendered,
			Ownership:   model.DeveloperOwned,
			Permissions: 0644,
		},
	}, nil
}
