package process

import (
	"context"
	"io"
)

// NoopRunner implements Runner by returning success (exit code 0) without spawning processes.
type NoopRunner struct{}

// NewNoopRunner constructs a NoopRunner.
func NewNoopRunner() *NoopRunner {
	return &NoopRunner{}
}

// Run returns an empty successful Result.
func (r *NoopRunner) Run(ctx context.Context, dir string, name string, args ...string) (*Result, error) {
	return &Result{ExitCode: 0}, nil
}

// RunWithInput returns an empty successful Result without consuming input.
func (r *NoopRunner) RunWithInput(ctx context.Context, dir string, in io.Reader, name string, args ...string) (*Result, error) {
	return &Result{ExitCode: 0}, nil
}
