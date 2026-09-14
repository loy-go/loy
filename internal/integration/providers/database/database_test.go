package database_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/loy-go/loy/internal/filesystem"
	"github.com/loy-go/loy/internal/integration/providers/database"
	"github.com/loy-go/loy/internal/process"
)

func TestDriverRegistry(t *testing.T) {
	reg := database.NewDriverRegistry()

	// Default postgres adapter should be available
	pg, err := reg.Get("postgres")
	if err != nil {
		t.Fatalf("expected postgres driver: %v", err)
	}
	if pg.Name() != "postgres" {
		t.Errorf("expected name postgres, got %s", pg.Name())
	}
	if pg.DriverName() != "pgx" {
		t.Errorf("expected driver pgx, got %s", pg.DriverName())
	}

	// Alias postgresql
	pgAlias, err := reg.Get("postgresql")
	if err != nil {
		t.Fatalf("expected postgresql alias driver: %v", err)
	}
	if pgAlias.DriverName() != "pgx" {
		t.Errorf("expected driver pgx, got %s", pgAlias.DriverName())
	}

	// Unknown driver
	if _, err := reg.Get("unknown"); err == nil {
		t.Errorf("expected error for unknown driver")
	}
}

func TestGooseRunnerCreate(t *testing.T) {
	memFS := filesystem.NewMemFileSystem()
	runner := database.NewGooseRunner(memFS, nil)

	ctx := context.Background()

	// Empty name should fail
	if _, err := runner.Create(ctx, "migrations", ""); err == nil {
		t.Errorf("expected error with empty migration name")
	}

	// Successful creation
	createdPath, err := runner.Create(ctx, "migrations", "create_users_table")
	if err != nil {
		t.Fatalf("unexpected error creating migration: %v", err)
	}

	if !strings.HasPrefix(filepath.Base(createdPath), "20") {
		t.Errorf("expected timestamp prefix in filename: %s", createdPath)
	}
	if !strings.HasSuffix(createdPath, "_create_users_table.sql") {
		t.Errorf("expected suffix _create_users_table.sql: %s", createdPath)
	}

	// Verify content
	content, err := memFS.ReadFile(createdPath)
	if err != nil {
		t.Fatalf("reading created migration file: %v", err)
	}
	sContent := string(content)
	if !strings.Contains(sContent, "-- +goose Up") || !strings.Contains(sContent, "-- +goose Down") {
		t.Errorf("migration file missing goose annotations:\n%s", sContent)
	}
}

func TestGooseRunnerSecurityValidation(t *testing.T) {
	memFS := filesystem.NewMemFileSystem()
	runner := database.NewGooseRunner(memFS, nil)
	ctx := context.Background()

	// 1. Path traversal in migration creation
	traversalNames := []string{"../escape", "foo/bar", "foo\\bar", "nested/name"}
	for _, name := range traversalNames {
		if _, err := runner.Create(ctx, "migrations", name); err == nil {
			t.Errorf("expected path traversal error for migration name %q", name)
		}
	}

	// 2. Table name SQL injection
	badTables := []string{
		"users; DROP TABLE users;--",
		"table name with spaces",
		"table-with-dashes",
		"123startwithnumber",
	}
	for _, table := range badTables {
		opts := database.DefaultMigrationOptions()
		opts.TableName = table
		opts.DSN = "postgres://fake/db"
		err := runner.Up(ctx, opts)
		if err == nil {
			t.Errorf("expected SQL injection validation error for table name %q", table)
		}
		if !strings.Contains(err.Error(), "invalid migration table name") {
			t.Errorf("expected 'invalid migration table name' error, got: %v", err)
		}
	}
}

func TestSqlcConfigAndRunner(t *testing.T) {
	memFS := filesystem.NewMemFileSystem()
	execRunner := process.NewExecRunner()
	runner := database.NewSqlcRunner(memFS, execRunner)

	ctx := context.Background()

	// Scaffold config
	opts := database.DefaultSqlcConfigOptions()
	yaml := database.GenerateSqlcYAML(opts)
	if !strings.Contains(yaml, `version: "2"`) || !strings.Contains(yaml, `sql_package: "pgx/v5"`) {
		t.Errorf("unexpected generated sqlc.yaml:\n%s", yaml)
	}

	if err := runner.ScaffoldConfig(ctx, "/app", opts); err != nil {
		t.Fatalf("scaffolding sqlc config: %v", err)
	}

	cfgExists, _ := memFS.Exists("/app/sqlc.yaml")
	if !cfgExists {
		t.Errorf("expected /app/sqlc.yaml to exist")
	}

	// Scaffold again - should be idempotent and not overwrite
	if err := runner.ScaffoldConfig(ctx, "/app", opts); err != nil {
		t.Fatalf("idempotent scaffolding failed: %v", err)
	}
}

func TestGooseRunnerExecutionSkipWhenNoDB(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("skipping live goose migration execution; DATABASE_URL not set")
	}

	runner := database.NewGooseRunner(nil, nil)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	opts := database.MigrationOptions{
		Driver:    "postgres",
		DSN:       dsn,
		Directory: "testdata/migrations",
		TableName: "goose_db_version_test",
		Timeout:   10 * time.Second,
		Quiet:     true,
	}

	// Verify status runs without panic against real test DB
	_ = runner.Status(ctx, opts)
}
