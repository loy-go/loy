package main_test

import (
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func buildBinary(t *testing.T) string {
	binPath := filepath.Join(t.TempDir(), "loy")
	cmd := exec.Command("go", "build", "-o", binPath, ".")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("failed to build binary: %v, output: %s", err, string(out))
	}
	return binPath
}

func TestCLI_ExitCodes(t *testing.T) {
	bin := buildBinary(t)

	t.Run("success exit code 0", func(t *testing.T) {
		cmd := exec.Command(bin, "version")
		err := cmd.Run()
		if err != nil {
			t.Fatalf("expected exit code 0 for version, got %v", err)
		}
	})

	t.Run("usage error exit code 2", func(t *testing.T) {
		cmd := exec.Command(bin, "nonexistent-command")
		err := cmd.Run()
		if err == nil {
			t.Fatal("expected non-zero exit code for invalid command")
		}
		exitErr, ok := err.(*exec.ExitError)
		if !ok {
			t.Fatalf("expected exec.ExitError, got %T", err)
		}
		if exitErr.ExitCode() != 2 {
			t.Fatalf("expected exit code 2 for usage error, got %d", exitErr.ExitCode())
		}
	})

	t.Run("usage error with json output", func(t *testing.T) {
		cmd := exec.Command(bin, "invalid-cmd", "--json")
		out, err := cmd.CombinedOutput()
		if err == nil {
			t.Fatal("expected error")
		}
		if !strings.Contains(string(out), `"code": "LOY-CLI-001"`) {
			t.Fatalf("expected JSON diagnostic in stderr, got: %s", string(out))
		}
	})
}
