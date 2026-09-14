package builtin

import (
	"context"
	"fmt"

	"github.com/uloydev/loy/internal/generator"
	"github.com/uloydev/loy/internal/generator/builtin/wiring"
	"github.com/uloydev/loy/internal/generator/model"
)

// FeatureGenerator bundles model, repository, service, handler, test and wiring artifacts.
type FeatureGenerator struct {
	modulePath string
}

// NewFeatureGenerator constructs FeatureGenerator.
func NewFeatureGenerator(modulePath string) *FeatureGenerator {
	return &FeatureGenerator{modulePath: modulePath}
}

func (g *FeatureGenerator) Name() string {
	return "feature"
}

func (g *FeatureGenerator) Generate(ctx context.Context, input generator.Input) ([]model.Artifact, error) {
	if input.Name == "" {
		return nil, fmt.Errorf("feature name is required")
	}
	if !ValidIdentifierRegex.MatchString(input.Name) {
		return nil, fmt.Errorf("invalid feature identifier %q: must match %s", input.Name, ValidIdentifierRegex.String())
	}

	var artifacts []model.Artifact

	// 1. Model
	modelGen := NewModelGenerator(g.modulePath)
	arts, err := modelGen.Generate(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("generating model: %w", err)
	}
	artifacts = append(artifacts, arts...)

	// 2. Repository
	repoGen := NewRepositoryGenerator(g.modulePath)
	arts, err = repoGen.Generate(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("generating repository: %w", err)
	}
	artifacts = append(artifacts, arts...)

	// 3. Service
	svcGen := NewServiceGenerator(g.modulePath)
	arts, err = svcGen.Generate(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("generating service: %w", err)
	}
	artifacts = append(artifacts, arts...)

	// 4. Handler
	handlerGen := NewHandlerGenerator(g.modulePath)
	arts, err = handlerGen.Generate(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("generating handler: %w", err)
	}
	artifacts = append(artifacts, arts...)

	// 5. Test
	testGen := NewTestGenerator(g.modulePath)
	arts, err = testGen.Generate(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("generating test: %w", err)
	}
	artifacts = append(artifacts, arts...)

	// 6. Wiring Splicing
	wiringArts := wiring.GenerateWiringArtifacts(input.Name, g.modulePath)
	artifacts = append(artifacts, wiringArts...)

	return artifacts, nil
}
