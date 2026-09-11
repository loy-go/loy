package builtin

import (
	"context"
	"fmt"

	"github.com/uloydev/loy/internal/generator"
	"github.com/uloydev/loy/internal/generator/model"
)

// ModelGenerator creates domain entity struct and validation.
type ModelGenerator struct {
	modulePath string
}

// NewModelGenerator constructs ModelGenerator.
func NewModelGenerator(modulePath string) *ModelGenerator {
	return &ModelGenerator{modulePath: modulePath}
}

func (g *ModelGenerator) Name() string {
	return "model"
}

func (g *ModelGenerator) Generate(ctx context.Context, input generator.Input) ([]model.Artifact, error) {
	if input.Name == "" {
		return nil, fmt.Errorf("model name is required")
	}

	rawFields := input.Args["fields"]
	var rawList []string
	if rawFields != "" {
		rawList = splitRawArgs(rawFields)
	}

	fields, err := ParseFields(rawList)
	if err != nil {
		return nil, err
	}

	data := NewBaseData(input.Name, g.modulePath, fields)
	tmpl, err := ReadTemplate("model.go.tmpl")
	if err != nil {
		return nil, err
	}

	renderer := GetRenderer()
	rendered, err := renderer.RenderGo(ctx, "model.go.tmpl", tmpl, data)
	if err != nil {
		return nil, err
	}

	path := fmt.Sprintf("internal/%s/domain/%s.go", data.FeaturePkg, data.Snake)
	return []model.Artifact{
		{
			Path:        path,
			Content:     rendered,
			Ownership:   model.DeveloperOwned,
			Permissions: 0644,
		},
	}, nil
}
