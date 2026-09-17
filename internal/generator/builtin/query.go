package builtin

import (
	"context"
	"fmt"

	"github.com/loy-go/loy/internal/generator"
	"github.com/loy-go/loy/internal/generator/model"
)

// QueryGenerator scaffolds a CQRS read query and handler projection.
type QueryGenerator struct {
	modulePath string
}

// NewQueryGenerator constructs QueryGenerator.
func NewQueryGenerator(modulePath string) *QueryGenerator {
	return &QueryGenerator{modulePath: modulePath}
}

func (g *QueryGenerator) Name() string {
	return "query"
}

func (g *QueryGenerator) Generate(ctx context.Context, input generator.Input) ([]model.Artifact, error) {
	if input.Name == "" {
		return nil, fmt.Errorf("query name is required")
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
	tmpl, err := ReadTemplateContext(ctx, "query.go.tmpl")
	if err != nil {
		return nil, fmt.Errorf("reading query template: %w", err)
	}

	renderer := GetRenderer()
	rendered, err := renderer.RenderGo(ctx, "query.go.tmpl", tmpl, data)
	if err != nil {
		return nil, fmt.Errorf("rendering query: %w", err)
	}

	path := fmt.Sprintf("internal/%s/query/%s_query.go", data.FeaturePkg, data.Snake)
	return []model.Artifact{
		{
			Path:        path,
			Content:     rendered,
			Ownership:   model.DeveloperOwned,
			Permissions: 0644,
		},
	}, nil
}
