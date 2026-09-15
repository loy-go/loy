package database

import (
	"context"
	"fmt"

	"github.com/loy-go/loy/internal/filesystem"
	"github.com/loy-go/loy/internal/process"
)

// SqlcConfigOptions contains properties to generate a standard sqlc.yaml configuration.
type SqlcConfigOptions struct {
	Engine      string // e.g. "postgresql", "sqlite", "mysql"
	PackageName string // e.g. "db" or "repository"
	PackageOut  string // e.g. "internal/platform/database/sqlc"
	SchemaPath  string // e.g. "migrations"
	QueriesPath string // e.g. "queries"
}

// DefaultSqlcConfigOptions returns standard Loy layout defaults.
func DefaultSqlcConfigOptions() SqlcConfigOptions {
	return SqlcConfigOptions{
		Engine:      "postgresql",
		PackageName: "db",
		PackageOut:  "internal/platform/database/sqlc",
		SchemaPath:  "migrations",
		QueriesPath: "queries",
	}
}

// GenerateSqlcYAML generates a standard sqlc.yaml file configured for the target database engine.
func GenerateSqlcYAML(opts SqlcConfigOptions) string {
	if opts.PackageName == "" {
		opts.PackageName = "db"
	}
	if opts.PackageOut == "" {
		opts.PackageOut = "internal/platform/database/sqlc"
	}
	if opts.SchemaPath == "" {
		opts.SchemaPath = "migrations"
	}
	if opts.QueriesPath == "" {
		opts.QueriesPath = "queries"
	}

	engine := opts.Engine
	if engine == "" {
		engine = "postgresql"
	}

	if engine == "sqlite" || engine == "sqlite3" {
		return fmt.Sprintf(`version: "2"
sql:
  - engine: "sqlite"
    schema: "%s"
    queries: "%s"
    gen:
      go:
        package: "%s"
        out: "%s"
        emit_json_tags: true
        emit_prepared_queries: false
        emit_interface: true
        emit_empty_slices: true
`, opts.SchemaPath, opts.QueriesPath, opts.PackageName, opts.PackageOut)
	}

	if engine == "mysql" {
		return fmt.Sprintf(`version: "2"
sql:
  - engine: "mysql"
    schema: "%s"
    queries: "%s"
    gen:
      go:
        package: "%s"
        out: "%s"
        emit_json_tags: true
        emit_prepared_queries: false
        emit_interface: true
        emit_empty_slices: true
`, opts.SchemaPath, opts.QueriesPath, opts.PackageName, opts.PackageOut)
	}

	return fmt.Sprintf(`version: "2"
sql:
  - engine: "postgresql"
    schema: "%s"
    queries: "%s"
    gen:
      go:
        package: "%s"
        out: "%s"
        sql_package: "pgx/v5"
        emit_json_tags: true
        emit_prepared_queries: false
        emit_interface: true
        emit_empty_slices: true
`, opts.SchemaPath, opts.QueriesPath, opts.PackageName, opts.PackageOut)
}

// SqlcRunner manages execution and verification of external sqlc CLI tool.
type SqlcRunner struct {
	fs     filesystem.FileSystem
	runner process.Runner
}

// NewSqlcRunner constructs a new SqlcRunner.
func NewSqlcRunner(fs filesystem.FileSystem, runner process.Runner) *SqlcRunner {
	if fs == nil {
		fs = filesystem.NewOSFileSystem()
	}
	if runner == nil {
		runner = process.NewExecRunner()
	}
	return &SqlcRunner{
		fs:     fs,
		runner: runner,
	}
}

// IsInstalled checks if sqlc binary is available on system PATH.
func (r *SqlcRunner) IsInstalled(ctx context.Context) bool {
	res, err := r.runner.Run(ctx, "", "sqlc", "version")
	return err == nil && res != nil && res.ExitCode == 0
}

// ScaffoldConfig writes sqlc.yaml if not already present.
func (r *SqlcRunner) ScaffoldConfig(ctx context.Context, targetDir string, opts SqlcConfigOptions) error {
	safePath, err := filesystem.CleanAndValidatePath(targetDir, "sqlc.yaml")
	if err != nil {
		return fmt.Errorf("validating sqlc.yaml target path: %w", err)
	}

	exists, err := r.fs.Exists(safePath)
	if err != nil {
		return fmt.Errorf("checking sqlc.yaml existence: %w", err)
	}
	if exists {
		return nil // do not overwrite existing configuration
	}

	content := GenerateSqlcYAML(opts)
	if err := r.fs.WriteFile(safePath, []byte(content), 0644); err != nil {
		return fmt.Errorf("writing %s: %w", safePath, err)
	}
	return nil
}

// Generate executes 'sqlc generate' in the target module directory.
func (r *SqlcRunner) Generate(ctx context.Context, targetDir string) error {
	res, err := r.runner.Run(ctx, targetDir, "sqlc", "generate")
	if err != nil {
		return fmt.Errorf("executing sqlc generate: %w", err)
	}
	if res.ExitCode != 0 {
		return fmt.Errorf("sqlc generate failed (exit code %d): %s", res.ExitCode, string(res.Stderr))
	}
	return nil
}
