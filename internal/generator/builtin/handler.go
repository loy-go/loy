package builtin

import (
	"context"
	"fmt"

	"github.com/uloydev/loy/internal/generator"
	"github.com/uloydev/loy/internal/generator/model"
)

// HandlerGenerator creates HTTP transport handler and routes.
type HandlerGenerator struct {
	modulePath string
}

// NewHandlerGenerator constructs HandlerGenerator.
func NewHandlerGenerator(modulePath string) *HandlerGenerator {
	return &HandlerGenerator{modulePath: modulePath}
}

func (g *HandlerGenerator) Name() string {
	return "handler"
}

func (g *HandlerGenerator) Generate(ctx context.Context, input generator.Input) ([]model.Artifact, error) {
	if input.Name == "" {
		return nil, fmt.Errorf("handler name is required")
	}

	data := NewBaseData(input.Name, g.modulePath, nil)
	tmpl, err := ReadTemplate("handler.go.tmpl")
	if err != nil {
		return nil, err
	}

	renderer := GetRenderer()
	rendered, err := renderer.RenderGo(ctx, "handler.go.tmpl", tmpl, data)
	if err != nil {
		return nil, err
	}

	return []model.Artifact{
		{
			Path:        fmt.Sprintf("internal/%s/transport/http/handler.go", data.FeaturePkg),
			Content:     rendered,
			Ownership:   model.DeveloperOwned,
			Permissions: 0644,
		},
	}, nil
}
