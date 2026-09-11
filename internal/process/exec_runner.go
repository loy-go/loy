package process

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"time"
)

// ExecRunner executes processes using the operating system's os/exec package.
type ExecRunner struct{}

// NewExecRunner creates a new ExecRunner instance.
func NewExecRunner() *ExecRunner {
	return &ExecRunner{}
}

func (r *ExecRunner) Run(ctx context.Context, dir string, name string, args ...string) (*Result, error) {
	return r.RunWithInput(ctx, dir, nil, name, args...)
}

func (r *ExecRunner) RunWithInput(ctx context.Context, dir string, in io.Reader, name string, args ...string) (*Result, error) {
	if name == "" {
		return nil, errors.New("process name/command cannot be empty")
	}

	start := time.Now()
	cmd := exec.CommandContext(ctx, name, args...)
	if dir != "" {
		cmd.Dir = dir
	}
	if in != nil {
		cmd.Stdin = in
	}

	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf

	err := cmd.Run()
	duration := time.Since(start)

	res := &Result{
		Stdout:   stdoutBuf.Bytes(),
		Stderr:   stderrBuf.Bytes(),
		Duration: duration,
	}

	if err != nil {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			res.ExitCode = exitErr.ExitCode()
			return res, nil
		}
		return nil, fmt.Errorf("process execution failed: %w", err)
	}

	res.ExitCode = 0
	return res, nil
}
