package process_test

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	"github.com/loy-go/loy/internal/process"
)

func TestExecRunner_RunSuccess(t *testing.T) {
	runner := process.NewExecRunner()
	ctx := context.Background()

	res, err := runner.Run(ctx, "", "echo", "hello", "loy")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !res.Success() {
		t.Fatalf("expected success, got exit code %d", res.ExitCode)
	}

	output := strings.TrimSpace(string(res.Stdout))
	if output != "hello loy" {
		t.Fatalf("expected 'hello loy', got %q", output)
	}
}

func TestExecRunner_RunWithInput(t *testing.T) {
	runner := process.NewExecRunner()
	ctx := context.Background()
	input := bytes.NewBufferString("sample input stream")

	res, err := runner.RunWithInput(ctx, "", input, "cat")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !res.Success() {
		t.Fatalf("expected success, got exit code %d", res.ExitCode)
	}

	if string(res.Stdout) != "sample input stream" {
		t.Fatalf("expected 'sample input stream', got %q", string(res.Stdout))
	}
}

func TestExecRunner_NonZeroExit(t *testing.T) {
	runner := process.NewExecRunner()
	ctx := context.Background()

	// Running false exits with code 1
	res, err := runner.Run(ctx, "", "false")
	if err != nil {
		t.Fatalf("expected no go error for regular non-zero exit, got %v", err)
	}

	if res.Success() || res.ExitCode != 1 {
		t.Fatalf("expected exit code 1, got %d", res.ExitCode)
	}
}

func TestExecRunner_Timeout(t *testing.T) {
	runner := process.NewExecRunner()
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_, err := runner.Run(ctx, "", "sleep", "2")
	if err == nil {
		t.Fatalf("expected context deadline error, got nil")
	}
}
