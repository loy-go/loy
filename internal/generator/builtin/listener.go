package builtin

import (
	"context"
	"fmt"

	"github.com/uloydev/loy/internal/generator"
	"github.com/uloydev/loy/internal/generator/model"
)

// ListenerGenerator creates event listener consumer.
type ListenerGenerator struct {
	modulePath string
}

// NewListenerGenerator constructs ListenerGenerator.
func NewListenerGenerator(modulePath string) *ListenerGenerator {
	return &ListenerGenerator{modulePath: modulePath}
}

func (g *ListenerGenerator) Name() string {
	return "listener"
}

func (g *ListenerGenerator) Generate(ctx context.Context, input generator.Input) ([]model.Artifact, error) {
	if input.Name == "" {
		return nil, fmt.Errorf("listener name is required")
	}

	data := NewBaseData(input.Name, g.modulePath, nil)
	tmpl, err := ReadTemplate("listener.go.tmpl")
	if err != nil {
		return nil, err
	}

	renderer := GetRenderer()
	rendered, err := renderer.RenderGo(ctx, "listener.go.tmpl", tmpl, data)
	if err != nil {
		return nil, err
	}

	return []model.Artifact{
		{
			Path:        fmt.Sprintf("internal/%s/listener/%s_listener.go", data.FeaturePkg, data.Snake),
			Content:     rendered,
			Ownership:   model.DeveloperOwned,
			Permissions: 0644,
		},
	}, nil
}
