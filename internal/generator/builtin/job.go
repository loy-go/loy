package builtin

import (
	"context"
	"fmt"

	"github.com/uloydev/loy/internal/generator"
	"github.com/uloydev/loy/internal/generator/model"
)

// JobGenerator creates Asynq background task payload and processor.
type JobGenerator struct {
	modulePath string
}

// NewJobGenerator constructs JobGenerator.
func NewJobGenerator(modulePath string) *JobGenerator {
	return &JobGenerator{modulePath: modulePath}
}

func (g *JobGenerator) Name() string {
	return "job"
}

func (g *JobGenerator) Generate(ctx context.Context, input generator.Input) ([]model.Artifact, error) {
	if input.Name == "" {
		return nil, fmt.Errorf("job name is required")
	}

	data := NewBaseData(input.Name, g.modulePath, nil)
	tmpl, err := ReadTemplate("job.go.tmpl")
	if err != nil {
		return nil, err
	}

	renderer := GetRenderer()
	rendered, err := renderer.RenderGo(ctx, "job.go.tmpl", tmpl, data)
	if err != nil {
		return nil, err
	}

	return []model.Artifact{
		{
			Path:        fmt.Sprintf("internal/%s/job/%s_job.go", data.FeaturePkg, data.Snake),
			Content:     rendered,
			Ownership:   model.DeveloperOwned,
			Permissions: 0644,
		},
	}, nil
}
