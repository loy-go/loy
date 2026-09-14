package plan

import (
	"context"
	"fmt"

	"github.com/loy-go/loy/internal/diagnostics"
	"github.com/loy-go/loy/internal/filesystem"
	"github.com/loy-go/loy/internal/generator"
	"github.com/loy-go/loy/internal/generator/model"
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

	plannedPaths := make(map[string]model.Ownership)
	plannedRegions := make(map[string]map[string]bool)

	for _, art := range artifacts {
		relPath, err := ResolveTargetPath(targetDir, art.Path)
		if err != nil {
			diag := diagnostics.NewError(
				diagnostics.CodeFSPathTraversal,
				fmt.Sprintf("path validation failed for %q: %v", art.Path, err),
			)
			diag.File = art.Path
			return nil, &diag
		}

		var opType OperationType

		if art.Ownership == model.MixedOwned && art.Region != "" {
			// Check for duplicate region splice inside same plan
			if regions, exists := plannedRegions[relPath]; exists && regions[art.Region] {
				diag := diagnostics.NewError(
					diagnostics.CodeGenConflict,
					fmt.Sprintf("duplicate splice targeted at region %q in %q within same plan", art.Region, art.Path),
				)
				diag.File = art.Path
				return nil, &diag
			}

			// Check if file exists on disk OR is being created earlier in this plan
			existsOnDisk, _ := b.fs.Exists(relPath)
			_, createdInPlan := plannedPaths[relPath]

			if existsOnDisk || createdInPlan {
				opType = OpSplice
				if plannedRegions[relPath] == nil {
					plannedRegions[relPath] = make(map[string]bool)
				}
				plannedRegions[relPath][art.Region] = true
			} else {
				// Cannot splice non-existent file without base creation
				opType, err = b.detector.Check(ctx, targetDir, art, opts)
				if err != nil {
					return nil, err
				}
			}
		} else {
			// Intra-plan collision check for whole-file artifacts
			if prevOwner, existsInPlan := plannedPaths[relPath]; existsInPlan && prevOwner != model.MixedOwned {
				diag := diagnostics.NewError(
					diagnostics.CodeGenConflict,
					fmt.Sprintf("intra-plan collision: multiple whole-file artifacts targeted at %q", art.Path),
				)
				diag.File = art.Path
				diag.Hint = "Ensure each generated artifact has a distinct path or uses managed comment regions"
				return nil, &diag
			}

			opType, err = b.detector.Check(ctx, targetDir, art, opts)
			if err != nil {
				return nil, err
			}
			plannedPaths[relPath] = art.Ownership
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
