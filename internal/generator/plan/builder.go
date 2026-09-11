package plan

import (
	"context"
	"fmt"

	"github.com/uloydev/loy/internal/filesystem"
	"github.com/uloydev/loy/internal/generator"
	"github.com/uloydev/loy/internal/generator/model"
)

// Builder coordinates generation artifacts into an executable Plan.
type Builder struct {
	fs       filesystem.FileSystem
	detector *ConflictDetector
}

// NewBuilder constructs a Plan Builder.
func NewBuilder(fs filesystem.FileSystem) *Builder {
	return &Builder{
		fs:       fs,
		detector: NewConflictDetector(fs),
	}
}

// Build translates artifacts into an atomic execution Plan after validating conflicts.
func (b *Builder) Build(ctx context.Context, targetDir string, artifacts []model.Artifact, opts generator.Options) (*Plan, error) {
	plan := &Plan{
		TargetDirectory: targetDir,
		Operations:      make([]Operation, 0, len(artifacts)),
	}

	for _, art := range artifacts {
		opType, err := b.detector.Check(ctx, targetDir, art, opts)
		if err != nil {
			return nil, err
		}

		op := Operation{
			Type:        opType,
			Path:        art.Path,
			Content:     art.Content,
			Permissions: art.Permissions,
			Ownership:   art.Ownership,
			Region:      art.Region,
			Detail:      fmt.Sprintf("%s %s (%s)", opType, art.Path, art.Ownership),
		}

		plan.Operations = append(plan.Operations, op)
	}

	return plan, nil
}
