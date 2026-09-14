package plan

import (
	"io/fs"

	"github.com/loy-go/loy/internal/generator/model"
)

// OperationType specifies the kind of filesystem operation.
type OperationType string

const (
	// OpCreate creates a brand new file on disk.
	OpCreate OperationType = "create"

	// OpOverwrite replaces an existing GeneratedOwned file (or with force).
	OpOverwrite OperationType = "overwrite"

	// OpSplice updates an identifiable managed region in a MixedOwned file.
	OpSplice OperationType = "splice"

	// OpSkip skips write because existing content is identical.
	OpSkip OperationType = "skip"
)

// Operation describes a planned mutation to a single path.
type Operation struct {
	Type        OperationType
	Path        string
	Content     []byte
	Permissions fs.FileMode
	Ownership   model.Ownership
	Region      string
	Detail      string
}

// Plan represents an immutable, validated sequence of operations.
type Plan struct {
	TargetDirectory string
	Operations      []Operation
}
