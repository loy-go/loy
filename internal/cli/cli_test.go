package cli_test

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"

	"github.com/uloydev/loy/internal/cli"
	"github.com/uloydev/loy/internal/version"
)

func TestVersionCmd_Output(t *testing.T) {
	cmd := cli.NewRootCmd()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"version"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error executing version cmd: %v", err)
	}
}

func TestDirectoryFlag(t *testing.T) {
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(origDir)
	})

	cmd := cli.NewRootCmd()
	cmd.SetArgs([]string{"--directory", "/tmp", "version"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error with --directory flag: %v", err)
	}
}

func TestVersion_Get(t *testing.T) {
	info := version.Get()
	if info.Version == "" {
		t.Fatal("expected non-empty version")
	}
	if info.GoVersion == "" {
		t.Fatal("expected non-empty go version")
	}

	data, err := json.Marshal(info)
	if err != nil {
		t.Fatalf("failed to marshal version info: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("marshaled version info is empty")
	}
}
