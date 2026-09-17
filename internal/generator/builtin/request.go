package builtin

import (
	"context"
	"fmt"

	"github.com/loy-go/loy/internal/generator"
	"github.com/loy-go/loy/internal/generator/model"
)

// RequestGenerator creates DTO request validation struct.
type RequestGenerator struct {
	modulePath string
}

// NewRequestGenerator constructs RequestGenerator.
func NewRequestGenerator(modulePath string) *RequestGenerator {
	return &RequestGenerator{modulePath: modulePath}
}

func (g *RequestGenerator) Name() string {
	return "request"
}

func (g *RequestGenerator) Generate(ctx context.Context, input generator.Input) ([]model.Artifact, error) {
	if input.Name == "" {
		return nil, fmt.Errorf("request name is required")
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
	tmpl, err := ReadTemplateContext(ctx, "request.go.tmpl")
	if err != nil {
		return nil, err
	}

	renderer := GetRenderer()
	rendered, err := renderer.RenderGo(ctx, "request.go.tmpl", tmpl, data)
	if err != nil {
		return nil, err
	}

	return []model.Artifact{
		{
			Path:        fmt.Sprintf("internal/%s/transport/http/request/%s_request.go", data.FeaturePkg, data.Snake),
			Content:     rendered,
			Ownership:   model.DeveloperOwned,
			Permissions: 0644,
		},
	}, nil
}
