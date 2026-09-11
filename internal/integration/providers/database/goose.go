package database

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/pressly/goose/v3"
	"github.com/uloydev/loy/internal/filesystem"
)

var validTableNameRegex = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)

// MigrationOptions defines settings for goose execution.
type MigrationOptions struct {
	Driver    string        // e.g. "postgres"
	DSN       string        // connection string
	Directory string        // path to migration files
	TableName string        // version tracking table name, default "goose_db_version"
	Timeout   time.Duration // context execution timeout, default 30s
	Stdout    io.Writer     // output stream for migration logs
	Stderr    io.Writer     // error stream for migration logs
	Quiet     bool          // suppress standard log output
}

// DefaultMigrationOptions provides sane production defaults.
func DefaultMigrationOptions() MigrationOptions {
	return MigrationOptions{
		Driver:    "postgres",
		Directory: "migrations",
		TableName: "goose_db_version",
		Timeout:   30 * time.Second,
		Stdout:    os.Stdout,
		Stderr:    os.Stderr,
	}
}

// MigrationStatusItem models a single migration's state.
type MigrationStatusItem struct {
	Version   int64     `json:"version"`
	AppliedAt string    `json:"applied_at"`
	State     string    `json:"state"` // "applied" or "pending"
	File      string    `json:"file"`
}

// GooseRunner encapsulates embedded goose migration operations.
type GooseRunner struct {
	fs             filesystem.FileSystem
	driverRegistry *DriverRegistry
}

// NewGooseRunner creates a new GooseRunner.
func NewGooseRunner(fs filesystem.FileSystem, drivers *DriverRegistry) *GooseRunner {
	if fs == nil {
		fs = filesystem.NewOSFileSystem()
	}
	if drivers == nil {
		drivers = NewDriverRegistry()
	}
	return &GooseRunner{
		fs:             fs,
		driverRegistry: drivers,
	}
}

// configureGoose applies driver, table, and logging settings safely.
func (r *GooseRunner) configureGoose(opts MigrationOptions) error {
	tableName := opts.TableName
	if tableName == "" {
		tableName = "goose_db_version"
	}

	if !validTableNameRegex.MatchString(tableName) {
		return fmt.Errorf("invalid migration table name %q: must match %s", tableName, validTableNameRegex.String())
	}

	goose.SetTableName(tableName)

	// Dialect selection
	dialect := "postgres"
	if opts.Driver == "pgx" || opts.Driver == "postgresql" {
		dialect = "postgres"
	} else if opts.Driver != "" {
		dialect = opts.Driver
	}

	if err := goose.SetDialect(dialect); err != nil {
		return fmt.Errorf("setting goose dialect %s: %w", dialect, err)
	}

	if opts.Quiet {
		goose.SetLogger(log.New(io.Discard, "", 0))
	} else if opts.Stdout != nil {
		goose.SetLogger(log.New(opts.Stdout, "", 0))
	}

	return nil
}

// openDB resolves the driver adapter and opens a sql.DB connection.
func (r *GooseRunner) openDB(opts MigrationOptions) (*sql.DB, error) {
	driverName := opts.Driver
	if driverName == "" {
		driverName = "postgres"
	}

	adapter, err := r.driverRegistry.Get(driverName)
	if err != nil {
		return nil, err
	}

	if opts.DSN == "" {
		return nil, fmt.Errorf("database connection string (DSN) cannot be empty")
	}

	return adapter.Open(opts.DSN)
}

// Up runs all pending migrations.
func (r *GooseRunner) Up(ctx context.Context, opts MigrationOptions) error {
	if err := r.configureGoose(opts); err != nil {
		return err
	}

	db, err := r.openDB(opts)
	if err != nil {
		return err
	}
	defer db.Close()

	timeout := opts.Timeout
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	runCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	return goose.UpContext(runCtx, db, opts.Directory)
}

// Down rolls back the latest migration batch.
func (r *GooseRunner) Down(ctx context.Context, opts MigrationOptions) error {
	if err := r.configureGoose(opts); err != nil {
		return err
	}

	db, err := r.openDB(opts)
	if err != nil {
		return err
	}
	defer db.Close()

	timeout := opts.Timeout
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	runCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	return goose.DownContext(runCtx, db, opts.Directory)
}

// DownTo rolls back migrations down to a specific version.
func (r *GooseRunner) DownTo(ctx context.Context, opts MigrationOptions, version int64) error {
	if err := r.configureGoose(opts); err != nil {
		return err
	}

	db, err := r.openDB(opts)
	if err != nil {
		return err
	}
	defer db.Close()

	timeout := opts.Timeout
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	runCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	return goose.DownToContext(runCtx, db, opts.Directory, version)
}

// Redo rolls back the latest migration and runs it again.
func (r *GooseRunner) Redo(ctx context.Context, opts MigrationOptions) error {
	if err := r.configureGoose(opts); err != nil {
		return err
	}

	db, err := r.openDB(opts)
	if err != nil {
		return err
	}
	defer db.Close()

	timeout := opts.Timeout
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	runCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	return goose.RedoContext(runCtx, db, opts.Directory)
}

// Reset rolls back all migrations.
func (r *GooseRunner) Reset(ctx context.Context, opts MigrationOptions) error {
	if err := r.configureGoose(opts); err != nil {
		return err
	}

	db, err := r.openDB(opts)
	if err != nil {
		return err
	}
	defer db.Close()

	timeout := opts.Timeout
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	runCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	return goose.ResetContext(runCtx, db, opts.Directory)
}

// Status outputs the status of all migrations.
func (r *GooseRunner) Status(ctx context.Context, opts MigrationOptions) error {
	if err := r.configureGoose(opts); err != nil {
		return err
	}

	db, err := r.openDB(opts)
	if err != nil {
		return err
	}
	defer db.Close()

	timeout := opts.Timeout
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	runCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	return goose.StatusContext(runCtx, db, opts.Directory)
}

// Version prints the current database version.
func (r *GooseRunner) Version(ctx context.Context, opts MigrationOptions) (int64, error) {
	if err := r.configureGoose(opts); err != nil {
		return 0, err
	}

	db, err := r.openDB(opts)
	if err != nil {
		return 0, err
	}
	defer db.Close()

	timeout := opts.Timeout
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	runCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	return goose.GetDBVersionContext(runCtx, db)
}

// Create generates a timestamped migration SQL file using the configured filesystem.
func (r *GooseRunner) Create(ctx context.Context, dir, name string) (string, error) {
	if name == "" {
		return "", fmt.Errorf("migration name cannot be empty")
	}
	if filepath.Base(name) != name || strings.ContainsAny(name, `/\`) || strings.Contains(name, "..") {
		return "", fmt.Errorf("invalid migration name %q: path traversal elements not allowed", name)
	}
	if dir == "" {
		dir = "migrations"
	}

	// Clean and validate target path
	timestamp := time.Now().UTC().Format("20060102150405")
	filename := fmt.Sprintf("%s_%s.sql", timestamp, name)

	// Ensure directory exists
	if err := r.fs.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("creating migrations directory %s: %w", dir, err)
	}

	filePath := filepath.Join(dir, filename)

	content := fmt.Sprintf(`-- +goose Up
-- +goose StatementBegin
-- SQL in this section is executed when the migration is applied.
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- SQL in this section is executed when the migration is rolled back.
-- +goose StatementEnd
`)

	if err := r.fs.WriteFile(filePath, []byte(content), 0644); err != nil {
		return "", fmt.Errorf("writing migration file %s: %w", filePath, err)
	}

	return filePath, nil
}
