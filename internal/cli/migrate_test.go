package cli_test

import (
	"bytes"
	"io/fs"
	"strings"
	"testing"

	"github.com/loy-go/loy/internal/cli"
	"github.com/loy-go/loy/internal/filesystem"
	"github.com/loy-go/loy/internal/process"
)

func TestMigrateCommandTree(t *testing.T) {
	memFS := filesystem.NewMemFileSystem()
	execRunner := process.NewExecRunner()

	rootCmd := cli.NewRootCmdWithFS(memFS, execRunner)

	// Subcommands under migrate
	subCommands := []string{"up", "down", "status", "create", "redo", "reset", "version"}
	migrateCmd, _, err := rootCmd.Find([]string{"migrate"})
	if err != nil || migrateCmd == nil {
		t.Fatalf("failed to find 'migrate' command: %v", err)
	}

	foundMap := make(map[string]bool)
	for _, sub := range migrateCmd.Commands() {
		foundMap[sub.Name()] = true
	}

	for _, expected := range subCommands {
		if !foundMap[expected] {
			t.Errorf("expected subcommand 'migrate %s' to exist", expected)
		}
	}
}

func TestMigrateCreateCommand(t *testing.T) {
	memFS := filesystem.NewMemFileSystem()
	execRunner := process.NewExecRunner()

	// 1. Text output
	outBuf := new(bytes.Buffer)
	rootCmd := cli.NewRootCmdWithFS(memFS, execRunner)
	rootCmd.SetOut(outBuf)
	rootCmd.SetArgs([]string{"migrate", "create", "add_orders_table", "--dir", "/testapp/migrations"})

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("unexpected error executing migrate create: %v", err)
	}

	outStr := outBuf.String()
	if !strings.Contains(outStr, "Created migration:") || !strings.Contains(outStr, "add_orders_table.sql") {
		t.Errorf("unexpected output from migrate create:\n%s", outStr)
	}

	// Verify file on filesystem via Walk
	var foundFiles []string
	_ = memFS.Walk("/testapp/migrations", func(path string, d fs.DirEntry, err error) error {
		if err == nil && !d.IsDir() {
			foundFiles = append(foundFiles, path)
		}
		return nil
	})

	if len(foundFiles) != 1 {
		t.Fatalf("expected 1 file in /testapp/migrations, got count: %d", len(foundFiles))
	}
	if !strings.HasSuffix(foundFiles[0], "_add_orders_table.sql") {
		t.Errorf("unexpected file name: %s", foundFiles[0])
	}

	// 2. JSON output
	jsonBuf := new(bytes.Buffer)
	jsonCmd := cli.NewRootCmdWithFS(memFS, execRunner)
	jsonCmd.SetOut(jsonBuf)
	jsonCmd.SetArgs([]string{"--json", "migrate", "create", "add_items_table", "--dir", "/testapp/migrations"})

	if err := jsonCmd.Execute(); err != nil {
		t.Fatalf("unexpected error executing migrate create --json: %v", err)
	}

	jsonStr := jsonBuf.String()
	if !strings.Contains(jsonStr, `"status"`) || !strings.Contains(jsonStr, `"created"`) || !strings.Contains(jsonStr, "add_items_table.sql") {
		t.Errorf("unexpected json output:\n%s", jsonStr)
	}
}

func TestMigrateDownValidation(t *testing.T) {
	memFS := filesystem.NewMemFileSystem()
	execRunner := process.NewExecRunner()

	rootCmd := cli.NewRootCmdWithFS(memFS, execRunner)
	rootCmd.SetArgs([]string{"migrate", "down", "--to", "-1", "--db-url", "postgres://localhost/test"})

	err := rootCmd.Execute()
	if err == nil {
		t.Errorf("expected error when --to is negative")
	}
	if !strings.Contains(err.Error(), "invalid rollback version") {
		t.Errorf("expected 'invalid rollback version' error, got: %v", err)
	}
}

func TestMigrateUpFailsWithoutDB(t *testing.T) {
	memFS := filesystem.NewMemFileSystem()
	execRunner := process.NewExecRunner()

	rootCmd := cli.NewRootCmdWithFS(memFS, execRunner)
	rootCmd.SetArgs([]string{"migrate", "up"})

	err := rootCmd.Execute()
	if err == nil {
		t.Errorf("expected migrate up to fail without database connection string")
	}
	if !strings.Contains(err.Error(), "database connection string required") {
		t.Errorf("expected 'database connection string required' error, got: %v", err)
	}
}
