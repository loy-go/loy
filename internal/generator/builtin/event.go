package builtin

import (
	"context"
	"fmt"

	"github.com/uloydev/loy/internal/generator"
	"github.com/uloydev/loy/internal/generator/model"
)

// EventGenerator creates domain event model.
type EventGenerator struct {
	modulePath string
}

// NewEventGenerator constructs EventGenerator.
func NewEventGenerator(modulePath string) *EventGenerator {
	return &EventGenerator{modulePath: modulePath}
}

func (g *EventGenerator) Name() string {
	return "event"
}

func (g *EventGenerator) Generate(ctx context.Context, input generator.Input) ([]model.Artifact, error) {
	if input.Name == "" {
		return nil, fmt.Errorf("event name is required")
	}

	data := NewBaseData(input.Name, g.modulePath, nil)
	tmpl, err := ReadTemplate("event.go.tmpl")
	if err != nil {
		return nil, err
	}

	renderer := GetRenderer()
	rendered, err := renderer.RenderGo(ctx, "event.go.tmpl", tmpl, data)
	if err != nil {
		return nil, err
	}

	return []model.Artifact{
		{
			Path:        fmt.Sprintf("internal/%s/event/%s_event.go", data.FeaturePkg, data.Snake),
			Content:     rendered,
			Ownership:   model.DeveloperOwned,
			Permissions: 0644,
		},
	}, nil
}
