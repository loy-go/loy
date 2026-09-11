package process

import (
	"context"
	"io"
	"time"
)

// Result captures the outcome of an external command execution.
type Result struct {
	ExitCode int           `json:"exit_code"`
	Stdout   []byte        `json:"stdout"`
	Stderr   []byte        `json:"stderr"`
	Duration time.Duration `json:"duration"`
}

// Success reports whether the command exited with return code 0.
func (r *Result) Success() bool {
	return r != nil && r.ExitCode == 0
}

// Runner defines the contract for executing operating system processes.
type Runner interface {
	// Run executes an executable with args in directory dir without shell interpolation.
	Run(ctx context.Context, dir string, name string, args ...string) (*Result, error)

	// RunWithInput executes an executable with args in directory dir, piping stdin.
	RunWithInput(ctx context.Context, dir string, in io.Reader, name string, args ...string) (*Result, error)
}
