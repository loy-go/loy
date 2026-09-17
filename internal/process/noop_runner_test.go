package process

import (
	"bytes"
	"context"
	"testing"
)

func TestNoopRunner(t *testing.T) {
	r := NewNoopRunner()
	ctx := context.Background()

	res, err := r.Run(ctx, ".", "dummy", "arg1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !res.Success() {
		t.Errorf("expected success, got exit code %d", res.ExitCode)
	}

	resInput, err := r.RunWithInput(ctx, ".", bytes.NewReader([]byte("data")), "dummy")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resInput.Success() {
		t.Errorf("expected success with input, got exit code %d", resInput.ExitCode)
	}
}
