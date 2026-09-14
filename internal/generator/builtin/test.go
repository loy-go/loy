package builtin

import (
	"context"
	"fmt"

	"github.com/loy-go/loy/internal/generator"
	"github.com/loy-go/loy/internal/generator/model"
)

// TestGenerator creates unit/integration test scaffold.
type TestGenerator struct {
	modulePath string
}

// NewTestGenerator constructs TestGenerator.
func NewTestGenerator(modulePath string) *TestGenerator {
	return &TestGenerator{modulePath: modulePath}
}

func (g *TestGenerator) Name() string {
	return "test"
}

func (g *TestGenerator) Generate(ctx context.Context, input generator.Input) ([]model.Artifact, error) {
	if input.Name == "" {
		return nil, fmt.Errorf("test target name is required")
	}

	data := NewBaseData(input.Name, g.modulePath, nil)
	tmpl, err := ReadTemplate("test.go.tmpl")
	if err != nil {
		return nil, err
	}

	renderer := GetRenderer()
	rendered, err := renderer.RenderGo(ctx, "test.go.tmpl", tmpl, data)
	if err != nil {
		return nil, err
	}

	return []model.Artifact{
		{
			Path:        fmt.Sprintf("internal/%s/service/service_test.go", data.FeaturePkg),
			Content:     rendered,
			Ownership:   model.DeveloperOwned,
			Permissions: 0644,
		},
	}, nil
}
