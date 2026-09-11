package plan

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/uloydev/loy/internal/diagnostics"
	"github.com/uloydev/loy/internal/filesystem"
	"github.com/uloydev/loy/internal/generator/splicer"
)

// JournalRecord captures previous state of a modified or created file.
type JournalRecord struct {
	Path           string
	ExistedBefore  bool
	OriginalData   []byte
	CreatedTempDir bool
}

// Executor applies an execution Plan atomically to a FileSystem.
type Executor struct {
	fs      filesystem.FileSystem
	splicer *splicer.Splicer
}

// NewExecutor constructs an Executor.
func NewExecutor(fs filesystem.FileSystem) *Executor {
	return &Executor{
		fs:      fs,
		splicer: splicer.New(),
	}
}

// Execute applies all operations in plan.
// If any operation fails, journal rolls back all mutations.
func (e *Executor) Execute(ctx context.Context, plan *Plan) error {
	var journal []JournalRecord

	for _, op := range plan.Operations {
		if op.Type == OpSkip {
			continue
		}

		validPath, err := ResolveTargetPath(plan.TargetDirectory, op.Path)
		if err != nil {
			e.rollback(ctx, journal)
			diag := diagnostics.NewError(
				diagnostics.CodeFSPathTraversal,
				fmt.Sprintf("path traversal detected for %q: %v", op.Path, err),
			)
			diag.File = op.Path
			return &diag
		}
		fullPath := validPath

		// Record current state for potential rollback
		existed, err := e.fs.Exists(fullPath)
		if err != nil {
			e.rollback(ctx, journal)
			return fmt.Errorf("checking existence for %s: %w", fullPath, err)
		}

		var orig []byte
		if existed {
			orig, err = e.fs.ReadFile(fullPath)
			if err != nil {
				e.rollback(ctx, journal)
				return fmt.Errorf("reading original content %s: %w", fullPath, err)
			}
		}

		rec := JournalRecord{
			Path:          fullPath,
			ExistedBefore: existed,
			OriginalData:  orig,
		}

		// Ensure parent directory exists
		dir := filepath.Dir(fullPath)
		if dir != "." && dir != "/" {
			dirExists, err := e.fs.Exists(dir)
			if err != nil {
				e.rollback(ctx, journal)
				return fmt.Errorf("checking dir %s: %w", dir, err)
			}
			if !dirExists {
				if err := e.fs.MkdirAll(dir, 0755); err != nil {
					e.rollback(ctx, journal)
					return fmt.Errorf("creating dir %s: %w", dir, err)
				}
			}
		}

		// Pre-register journal record prior to file mutation for reliable rollback
		journal = append(journal, rec)

		switch op.Type {
		case OpCreate, OpOverwrite:
			perm := op.Permissions
			if perm == 0 {
				perm = 0644
			}
			if err := e.fs.WriteFile(fullPath, op.Content, perm); err != nil {
				e.rollback(ctx, journal)
				return diagnostics.NewError(
					diagnostics.CodeGenExecutionError,
					fmt.Sprintf("writing file %s: %v", fullPath, err),
				)
			}

		case OpSplice:
			isGo := strings.HasSuffix(fullPath, ".go")
			spliced, err := e.splicer.SpliceRegion(orig, op.Region, string(op.Content), isGo)
			if err != nil {
				e.rollback(ctx, journal)
				return err
			}

			perm := op.Permissions
			if perm == 0 {
				perm = 0644
			}
			if err := e.fs.WriteFile(fullPath, spliced, perm); err != nil {
				e.rollback(ctx, journal)
				return diagnostics.NewError(
					diagnostics.CodeGenExecutionError,
					fmt.Sprintf("writing spliced file %s: %v", fullPath, err),
				)
			}
		}
	}

	return nil
}

// rollback restores filesystem state in reverse order.
func (e *Executor) rollback(ctx context.Context, journal []JournalRecord) {
	for i := len(journal) - 1; i >= 0; i-- {
		rec := journal[i]
		if !rec.ExistedBefore {
			_ = e.fs.Remove(rec.Path)
		} else {
			_ = e.fs.WriteFile(rec.Path, rec.OriginalData, 0644)
		}
	}
}
