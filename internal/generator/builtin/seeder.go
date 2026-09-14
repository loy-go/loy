package builtin

import (
	"context"
	"fmt"

	"github.com/loy-go/loy/internal/generator"
	"github.com/loy-go/loy/internal/generator/model"
)

// SeederGenerator creates database seeder files.
type SeederGenerator struct {
	modulePath string
}

// NewSeederGenerator constructs SeederGenerator.
func NewSeederGenerator(modulePath string) *SeederGenerator {
	return &SeederGenerator{modulePath: modulePath}
}

func (g *SeederGenerator) Name() string {
	return "seeder"
}

func (g *SeederGenerator) Generate(ctx context.Context, input generator.Input) ([]model.Artifact, error) {
	if input.Name == "" {
		return nil, fmt.Errorf("seeder name is required")
	}

	data := NewBaseData(input.Name, g.modulePath, nil)
	renderer := GetRenderer()

	tmpl, err := ReadTemplate("seeder.go.tmpl")
	if err != nil {
		return nil, fmt.Errorf("reading seeder template: %w", err)
	}
	rendered, err := renderer.RenderGo(ctx, "seeder.go.tmpl", tmpl, data)
	if err != nil {
		return nil, fmt.Errorf("rendering seeder: %w", err)
	}

	return []model.Artifact{
		{
			Path:        fmt.Sprintf("internal/platform/database/seeds/%s_seeder.go", data.Snake),
			Content:     rendered,
			Ownership:   model.DeveloperOwned,
			Permissions: 0644,
		},
	}, nil
}
