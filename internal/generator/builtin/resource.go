package builtin

import (
	"context"
	"fmt"

	"github.com/uloydev/loy/internal/generator"
	"github.com/uloydev/loy/internal/generator/model"
)

// ResourceGenerator creates DTO response transform struct.
type ResourceGenerator struct {
	modulePath string
}

// NewResourceGenerator constructs ResourceGenerator.
func NewResourceGenerator(modulePath string) *ResourceGenerator {
	return &ResourceGenerator{modulePath: modulePath}
}

func (g *ResourceGenerator) Name() string {
	return "resource"
}

func (g *ResourceGenerator) Generate(ctx context.Context, input generator.Input) ([]model.Artifact, error) {
	if input.Name == "" {
		return nil, fmt.Errorf("resource name is required")
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
	tmpl, err := ReadTemplate("resource.go.tmpl")
	if err != nil {
		return nil, err
	}

	renderer := GetRenderer()
	rendered, err := renderer.RenderGo(ctx, "resource.go.tmpl", tmpl, data)
	if err != nil {
		return nil, err
	}

	return []model.Artifact{
		{
			Path:        fmt.Sprintf("internal/%s/transport/http/resource/%s_resource.go", data.FeaturePkg, data.Snake),
			Content:     rendered,
			Ownership:   model.DeveloperOwned,
			Permissions: 0644,
		},
	}, nil
}
