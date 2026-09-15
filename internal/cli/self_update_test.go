package cli_test

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/loy-go/loy/internal/cli"
	"github.com/loy-go/loy/internal/filesystem"
	"github.com/loy-go/loy/internal/process"
)

func TestSelfUpdateCommand_Help(t *testing.T) {
	memFS := filesystem.NewMemFileSystem()
	execRunner := process.NewExecRunner()

	buf := new(bytes.Buffer)
	rootCmd := cli.NewRootCmdWithFS(memFS, execRunner)
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"self-update", "--help"})

	err := rootCmd.ExecuteContext(context.Background())
	if err != nil {
		t.Fatalf("self-update --help failed: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "loy self-update checks GitHub Releases") || !strings.Contains(out, "--check-only") {
		t.Errorf("expected self-update help message, got: %s", out)
	}
}
