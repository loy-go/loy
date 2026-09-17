package builtin

import (
	"context"
	"fmt"

	"github.com/loy-go/loy/internal/generator"
	"github.com/loy-go/loy/internal/generator/model"
)

// CommandGenerator scaffolds a CQRS write command and handler pipeline.
type CommandGenerator struct {
	modulePath string
}

// NewCommandGenerator constructs CommandGenerator.
func NewCommandGenerator(modulePath string) *CommandGenerator {
	return &CommandGenerator{modulePath: modulePath}
}

func (g *CommandGenerator) Name() string {
	return "command"
}

func (g *CommandGenerator) Generate(ctx context.Context, input generator.Input) ([]model.Artifact, error) {
	if input.Name == "" {
		return nil, fmt.Errorf("command name is required")
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
	tmpl, err := ReadTemplateContext(ctx, "command.go.tmpl")
	if err != nil {
		return nil, fmt.Errorf("reading command template: %w", err)
	}

	renderer := GetRenderer()
	rendered, err := renderer.RenderGo(ctx, "command.go.tmpl", tmpl, data)
	if err != nil {
		return nil, fmt.Errorf("rendering command: %w", err)
	}

	path := fmt.Sprintf("internal/%s/command/%s_cmd.go", data.FeaturePkg, data.Snake)
	return []model.Artifact{
		{
			Path:        path,
			Content:     rendered,
			Ownership:   model.DeveloperOwned,
			Permissions: 0644,
		},
	}, nil
}
