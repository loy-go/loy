package cli_test

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/loy-go/loy/internal/cli"
	"github.com/loy-go/loy/internal/filesystem"
	"github.com/loy-go/loy/internal/mcp"
	"github.com/loy-go/loy/internal/process"
)

func TestMCPCommand_Execution(t *testing.T) {
	memFS := filesystem.NewMemFileSystem()
	runner := process.NewExecRunner()

	_ = memFS.MkdirAll("/proj", 0755)
	_ = memFS.WriteFile("/proj/go.mod", []byte("module github.com/test/mcpapp\n\ngo 1.22\n"), 0644)
	_ = memFS.WriteFile("/proj/loy.yaml", []byte("version: 1\nproject:\n  name: mcpapp\n"), 0644)

	input := `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05"}}` + "\n"
	inBuf := strings.NewReader(input)
	var outBuf bytes.Buffer

	rootCmd := cli.NewRootCmdWithFS(memFS, runner)
	rootCmd.SetIn(inBuf)
	rootCmd.SetOut(&outBuf)
	rootCmd.SetArgs([]string{"mcp", "/proj"})

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := rootCmd.ExecuteContext(ctx)
	if err != nil {
		t.Fatalf("mcp execution failed: %v", err)
	}

	out := strings.TrimSpace(outBuf.String())
	var resp mcp.JSONRPCResponse
	if err := json.Unmarshal([]byte(out), &resp); err != nil {
		t.Fatalf("failed to parse JSON-RPC response: %v\nOutput: %s", err, out)
	}

	if resp.ID != float64(1) {
		t.Errorf("expected ID 1, got %v", resp.ID)
	}
	if resp.Error != nil {
		t.Errorf("unexpected error in response: %+v", resp.Error)
	}
}
