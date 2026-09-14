package builtin

import (
	"context"
	"fmt"

	"github.com/loy-go/loy/internal/generator"
	"github.com/loy-go/loy/internal/generator/model"
)

// RepositoryGenerator creates domain repository interface and pg adapter.
type RepositoryGenerator struct {
	modulePath string
}

// NewRepositoryGenerator constructs RepositoryGenerator.
func NewRepositoryGenerator(modulePath string) *RepositoryGenerator {
	return &RepositoryGenerator{modulePath: modulePath}
}

func (g *RepositoryGenerator) Name() string {
	return "repository"
}

func (g *RepositoryGenerator) Generate(ctx context.Context, input generator.Input) ([]model.Artifact, error) {
	if input.Name == "" {
		return nil, fmt.Errorf("repository name is required")
	}

	data := NewBaseData(input.Name, g.modulePath, nil)
	renderer := GetRenderer()

	// 1. Interface artifact in domain/
	ifaceTmpl, err := ReadTemplate("repository_interface.go.tmpl")
	if err != nil {
		return nil, err
	}
	ifaceRendered, err := renderer.RenderGo(ctx, "repository_interface.go.tmpl", ifaceTmpl, data)
	if err != nil {
		return nil, err
	}

	// 2. Adapter artifact in repository/
	adapterTmpl, err := ReadTemplate("repository_adapter.go.tmpl")
	if err != nil {
		return nil, err
	}
	adapterRendered, err := renderer.RenderGo(ctx, "repository_adapter.go.tmpl", adapterTmpl, data)
	if err != nil {
		return nil, err
	}

	return []model.Artifact{
		{
			Path:        fmt.Sprintf("internal/%s/domain/repository.go", data.FeaturePkg),
			Content:     ifaceRendered,
			Ownership:   model.DeveloperOwned,
			Permissions: 0644,
		},
		{
			Path:        fmt.Sprintf("internal/%s/repository/pg_adapter.go", data.FeaturePkg),
			Content:     adapterRendered,
			Ownership:   model.DeveloperOwned,
			Permissions: 0644,
		},
	}, nil
}
