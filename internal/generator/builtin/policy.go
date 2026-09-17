package builtin

import (
	"context"
	"fmt"

	"github.com/loy-go/loy/internal/generator"
	"github.com/loy-go/loy/internal/generator/model"
)

// PolicyGenerator creates authorization policy.
type PolicyGenerator struct {
	modulePath string
}

// NewPolicyGenerator constructs PolicyGenerator.
func NewPolicyGenerator(modulePath string) *PolicyGenerator {
	return &PolicyGenerator{modulePath: modulePath}
}

func (g *PolicyGenerator) Name() string {
	return "policy"
}

func (g *PolicyGenerator) Generate(ctx context.Context, input generator.Input) ([]model.Artifact, error) {
	if input.Name == "" {
		return nil, fmt.Errorf("policy name is required")
	}

	data := NewBaseData(input.Name, g.modulePath, nil)
	tmpl, err := ReadTemplateContext(ctx, "policy.go.tmpl")
	if err != nil {
		return nil, err
	}

	renderer := GetRenderer()
	rendered, err := renderer.RenderGo(ctx, "policy.go.tmpl", tmpl, data)
	if err != nil {
		return nil, err
	}

	return []model.Artifact{
		{
			Path:        fmt.Sprintf("internal/%s/policy/%s_policy.go", data.FeaturePkg, data.Snake),
			Content:     rendered,
			Ownership:   model.DeveloperOwned,
			Permissions: 0644,
		},
	}, nil
}
