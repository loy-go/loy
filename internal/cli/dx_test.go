package cli_test

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/loy-go/loy/internal/cli"
	"github.com/loy-go/loy/internal/filesystem"
	"github.com/loy-go/loy/internal/process"
)

func TestDXCommandsTree(t *testing.T) {
	memFS := filesystem.NewMemFileSystem()
	execRunner := process.NewExecRunner()

	rootCmd := cli.NewRootCmdWithFS(memFS, execRunner)

	expectedCmds := []string{"dev", "doctor", "graph", "completion", "upgrade"}
	foundMap := make(map[string]bool)
	for _, cmd := range rootCmd.Commands() {
		foundMap[cmd.Name()] = true
	}

	for _, name := range expectedCmds {
		if !foundMap[name] {
			t.Errorf("expected command %q to be registered in root", name)
		}
	}
}

func TestDoctorCommandOutput(t *testing.T) {
	memFS := filesystem.NewMemFileSystem()
	execRunner := process.NewExecRunner()

	// 1. Text output
	buf := new(bytes.Buffer)
	rootCmd := cli.NewRootCmdWithFS(memFS, execRunner)
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"doctor", "/tmp"})

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("doctor execution failed: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "Go Compiler") {
		t.Errorf("expected 'Go Compiler' in doctor output:\n%s", out)
	}

	// 2. JSON output
	jsonBuf := new(bytes.Buffer)
	jsonCmd := cli.NewRootCmdWithFS(memFS, execRunner)
	jsonCmd.SetOut(jsonBuf)
	jsonCmd.SetArgs([]string{"--json", "doctor", "/tmp"})

	if err := jsonCmd.Execute(); err != nil {
		t.Fatalf("doctor --json execution failed: %v", err)
	}

	jsonOut := jsonBuf.String()
	if !strings.Contains(jsonOut, `"name":"Go Compiler"`) && !strings.Contains(jsonOut, `"name": "Go Compiler"`) {
		t.Errorf("expected json output with Go Compiler check:\n%s", jsonOut)
	}
}

func TestGraphCommandFormats(t *testing.T) {
	memFS := filesystem.NewMemFileSystem()
	execRunner := process.NewExecRunner()

	// Setup small valid module
	_ = memFS.WriteFile("/pkg/go.mod", []byte("module github.com/test/pkg\n\ngo 1.22\n"), 0644)
	_ = memFS.MkdirAll("/pkg/internal/domain", 0755)
	_ = memFS.WriteFile("/pkg/internal/domain/entity.go", []byte("package domain\n"), 0644)

	// ASCII format
	asciiBuf := new(bytes.Buffer)
	rootCmd := cli.NewRootCmdWithFS(memFS, execRunner)
	rootCmd.SetOut(asciiBuf)
	rootCmd.SetArgs([]string{"graph", "/pkg", "--format=ascii"})

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("graph --format=ascii failed: %v", err)
	}

	// Mermaid format
	mBuf := new(bytes.Buffer)
	mCmd := cli.NewRootCmdWithFS(memFS, execRunner)
	mCmd.SetOut(mBuf)
	mCmd.SetArgs([]string{"graph", "/pkg", "--format=mermaid"})

	if err := mCmd.Execute(); err != nil {
		t.Fatalf("graph --format=mermaid failed: %v", err)
	}

	if !strings.Contains(mBuf.String(), "```mermaid") {
		t.Errorf("expected mermaid block in output:\n%s", mBuf.String())
	}
}

func TestCompletionCommand(t *testing.T) {
	memFS := filesystem.NewMemFileSystem()
	execRunner := process.NewExecRunner()

	buf := new(bytes.Buffer)
	rootCmd := cli.NewRootCmdWithFS(memFS, execRunner)
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"completion", "bash"})

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("completion bash failed: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "bash completion") && !strings.Contains(out, "__loy") {
		t.Errorf("expected bash completion script in buffer, got: %s", out)
	}
}

func TestGraphViolationsOnly(t *testing.T) {
	memFS := filesystem.NewMemFileSystem()
	execRunner := process.NewExecRunner()

	// Setup module with architectural violation: domain imports infrastructure
	_ = memFS.WriteFile("/badpkg/go.mod", []byte("module github.com/test/badpkg\n\ngo 1.22\n"), 0644)
	_ = memFS.MkdirAll("/badpkg/internal/domain", 0755)
	_ = memFS.MkdirAll("/badpkg/internal/infrastructure/db", 0755)
	_ = memFS.WriteFile("/badpkg/internal/infrastructure/db/db.go", []byte("package db\n"), 0644)
	_ = memFS.WriteFile("/badpkg/internal/domain/user.go", []byte("package domain\n\nimport \"github.com/test/badpkg/internal/infrastructure/db\"\n\nvar _ = db.Any\n"), 0644)

	buf := new(bytes.Buffer)
	rootCmd := cli.NewRootCmdWithFS(memFS, execRunner)
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"graph", "/badpkg", "--format=ascii", "--violations-only"})

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("graph --violations-only failed: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "(VIOLATION)") {
		t.Errorf("expected (VIOLATION) tag in --violations-only output, got:\n%s", out)
	}
}

func TestDoctorStrictFailure(t *testing.T) {
	memFS := filesystem.NewMemFileSystem()
	execRunner := process.NewExecRunner()

	// Empty dir has warnings (missing go.mod / loy.yaml), --strict should fail with CommandError
	buf := new(bytes.Buffer)
	rootCmd := cli.NewRootCmdWithFS(memFS, execRunner)
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"doctor", "/empty", "--strict"})

	err := rootCmd.Execute()
	if err == nil {
		t.Fatalf("expected doctor --strict to return error on warnings")
	}

	var cmdErr *cli.CommandError
	if !errors.As(err, &cmdErr) {
		t.Errorf("expected *cli.CommandError, got: %T (%v)", err, err)
	} else if cmdErr.Code != 1 {
		t.Errorf("expected exit code 1, got: %d", cmdErr.Code)
	}
}

func TestGraphDiffCommand(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping git-based graph diff test in short mode")
	}

	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(origDir) }()

	tmpDir := t.TempDir()
	_ = os.Chdir(tmpDir)

	execRunner := process.NewExecRunner()
	ctx := context.Background()

	// Init git repo
	_, err = execRunner.Run(ctx, tmpDir, "git", "init", "-b", "main")
	if err != nil {
		t.Fatalf("git init: %v", err)
	}
	_, _ = execRunner.Run(ctx, tmpDir, "git", "config", "user.name", "Test")
	_, _ = execRunner.Run(ctx, tmpDir, "git", "config", "user.email", "test@example.com")

	// Base commit
	_ = os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte("module github.com/test/diffapp\n\ngo 1.22\n"), 0644)
	_ = os.MkdirAll(filepath.Join(tmpDir, "internal/domain"), 0755)
	_ = os.WriteFile(filepath.Join(tmpDir, "internal/domain/user.go"), []byte("package domain\ntype User struct{}\n"), 0644)

	_, _ = execRunner.Run(ctx, tmpDir, "git", "add", ".")
	_, _ = execRunner.Run(ctx, tmpDir, "git", "commit", "-m", "initial commit")

	// Now add a new service package that imports domain
	_ = os.MkdirAll(filepath.Join(tmpDir, "internal/service"), 0755)
	_ = os.WriteFile(filepath.Join(tmpDir, "internal/service/user.go"), []byte(`package service
import "github.com/test/diffapp/internal/domain"
type Svc struct { U domain.User }
`), 0644)

	// Test graph --diff main in ascii format
	buf := new(bytes.Buffer)
	rootCmd := cli.NewRootCmdWithFS(filesystem.NewOSFileSystem(), execRunner)
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"graph", tmpDir, "--diff", "main", "--format", "ascii"})

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("graph --diff main failed: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "Added Packages:") || !strings.Contains(out, "internal/service") {
		t.Errorf("expected added internal/service in diff output:\n%s", out)
	}

	// Test graph --diff main in markdown format
	mdBuf := new(bytes.Buffer)
	mdCmd := cli.NewRootCmdWithFS(filesystem.NewOSFileSystem(), execRunner)
	mdCmd.SetOut(mdBuf)
	mdCmd.SetArgs([]string{"graph", tmpDir, "--diff", "main", "--format", "markdown"})

	if err := mdCmd.Execute(); err != nil {
		t.Fatalf("graph --diff main markdown failed: %v", err)
	}

	mdOut := mdBuf.String()
	if !strings.Contains(mdOut, "Architecture Drift Report") || !strings.Contains(mdOut, "No new architectural violations introduced") {
		t.Errorf("expected clean architecture report in markdown:\n%s", mdOut)
	}
}
