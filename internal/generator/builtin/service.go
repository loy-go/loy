package builtin

import (
	"context"
	"fmt"

	"github.com/loy-go/loy/internal/generator"
	"github.com/loy-go/loy/internal/generator/model"
)

// ServiceGenerator creates application use cases coordinating domain entities.
type ServiceGenerator struct {
	modulePath string
}

// NewServiceGenerator constructs ServiceGenerator.
func NewServiceGenerator(modulePath string) *ServiceGenerator {
	return &ServiceGenerator{modulePath: modulePath}
}

func (g *ServiceGenerator) Name() string {
	return "service"
}

func (g *ServiceGenerator) Generate(ctx context.Context, input generator.Input) ([]model.Artifact, error) {
	if input.Name == "" {
		return nil, fmt.Errorf("service name is required")
	}

	data := NewBaseData(input.Name, g.modulePath, nil)
	tmpl, err := ReadTemplateContext(ctx, "service.go.tmpl")
	if err != nil {
		return nil, err
	}

	renderer := GetRenderer()
	rendered, err := renderer.RenderGo(ctx, "service.go.tmpl", tmpl, data)
	if err != nil {
		return nil, err
	}

	return []model.Artifact{
		{
			Path:        fmt.Sprintf("internal/%s/service/service.go", data.FeaturePkg),
			Content:     rendered,
			Ownership:   model.DeveloperOwned,
			Permissions: 0644,
		},
	}, nil
}
