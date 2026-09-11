package generator

import (
	"context"

	"github.com/uloydev/loy/internal/generator/model"
)

// Options carries generator runtime flags.
type Options struct {
	Force  bool
	DryRun bool
}

// Input encapsulates the generator input parameters.
type Input struct {
	Name    string
	Args    map[string]string
	Options Options
}

// Generator is the interface implemented by all Loy code generators.
type Generator interface {
	Name() string
	Generate(ctx context.Context, input Input) ([]model.Artifact, error)
}
