package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/spf13/cobra"
	"github.com/loy-go/loy/internal/discovery"
	"github.com/loy-go/loy/internal/filesystem"
	"github.com/loy-go/loy/internal/integration/providers/database"
	"github.com/loy-go/loy/internal/manifest"
	"github.com/loy-go/loy/internal/process"
)

type migrateFlags struct {
	driver    string
	dbURL     string
	dir       string
	table     string
	timeout   time.Duration
}

func newMigrateCmd(fs filesystem.FileSystem, runner process.Runner) *cobra.Command {
	flags := &migrateFlags{}

	cmd := &cobra.Command{
		Use:   "migrate",
		Short: "Manage database schema migrations via embedded goose engine",
		Long:  `loy migrate provides zero-install database schema evolution powered by an embedded goose migration engine.`,
	}

	cmd.PersistentFlags().StringVar(&flags.driver, "driver", "", "Database driver (postgres, default from loy.yaml or postgres)")
	cmd.PersistentFlags().StringVar(&flags.dbURL, "db-url", "", "Database connection URL (or DATABASE_URL env var)")
	cmd.PersistentFlags().StringVar(&flags.dir, "dir", "", "Migrations directory (default 'migrations')")
	cmd.PersistentFlags().StringVar(&flags.table, "table", "", "Goose db version table (default 'goose_db_version')")
	cmd.PersistentFlags().DurationVar(&flags.timeout, "timeout", 30*time.Second, "Execution timeout for migration operations")

	gooseRunner := database.NewGooseRunner(fs, nil)

	cmd.AddCommand(newMigrateUpCmd(fs, runner, flags, gooseRunner))
	cmd.AddCommand(newMigrateDownCmd(fs, runner, flags, gooseRunner))
	cmd.AddCommand(newMigrateStatusCmd(fs, runner, flags, gooseRunner))
	cmd.AddCommand(newMigrateCreateCmd(fs, runner, flags, gooseRunner))
	cmd.AddCommand(newMigrateRedoCmd(fs, runner, flags, gooseRunner))
	cmd.AddCommand(newMigrateResetCmd(fs, runner, flags, gooseRunner))
	cmd.AddCommand(newMigrateVersionCmd(fs, runner, flags, gooseRunner))

	return cmd
}

// resolveOptions computes migration settings following the priority cascade:
// CLI flags > Environment variables > loy.yaml manifest > Defaults.
func resolveMigrationOptions(ctx context.Context, fs filesystem.FileSystem, runner process.Runner, flags *migrateFlags, cmd *cobra.Command) (database.MigrationOptions, error) {
	opts := database.DefaultMigrationOptions()
	opts.Driver = "" // reset so cascade can evaluate loy.yaml fallback
	cliOpts := GetOptions(ctx)

	opts.Stdout = cmd.OutOrStdout()
	opts.Stderr = cmd.ErrOrStderr()
	opts.Quiet = cliOpts.Quiet || cliOpts.JSON

	if flags.timeout > 0 {
		opts.Timeout = flags.timeout
	}

	// 1. Directory
	if flags.dir != "" {
		opts.Directory = flags.dir
	} else if envDir := os.Getenv("GOOSE_MIGRATION_DIR"); envDir != "" {
		opts.Directory = envDir
	}

	// 2. Table
	if flags.table != "" {
		opts.TableName = flags.table
	}

	// 3. Driver & DSN from flags and env
	if flags.driver != "" {
		opts.Driver = flags.driver
	}
	if flags.dbURL != "" {
		opts.DSN = flags.dbURL
	} else if envURL := os.Getenv("DATABASE_URL"); envURL != "" {
		opts.DSN = envURL
	}

	// 4. Try resolving from loy.yaml manifest only if driver or DSN still missing
	if opts.Driver == "" || opts.DSN == "" {
		cwd, err := os.Getwd()
		if err == nil {
			disc, err := discovery.NewDiscoverer(fs, runner)
			if err == nil {
				res, diag := disc.Discover(ctx, cwd)
				if diag == nil && res != nil && res.HasManifest {
					mPath := res.ManifestPath
					if data, err := fs.ReadFile(mPath); err == nil {
						mParser := manifest.NewParser()
						man, mDiag := mParser.ParseStrict(mPath, data)
						if mDiag == nil && man != nil {
							if opts.Driver == "" && man.Defaults.Database != "" {
								opts.Driver = man.Defaults.Database
							}
							if opts.DSN == "" {
								if dbInt, ok := man.Integrations["database"]; ok {
									if dsnVal, ok := dbInt.Config["dsn"].(string); ok && dsnVal != "" {
										opts.DSN = dsnVal
									}
								}
							}
						}
					}
				}
			}
		}
	}

	// Sanity checks
	if opts.Driver == "" {
		opts.Driver = "postgres"
	}
	if opts.Directory == "" {
		opts.Directory = "migrations"
	}

	return opts, nil
}

func newMigrateUpCmd(fs filesystem.FileSystem, runner process.Runner, flags *migrateFlags, gRunner *database.GooseRunner) *cobra.Command {
	return &cobra.Command{
		Use:   "up",
		Short: "Apply all pending database migrations",
		RunE: func(cmd *cobra.Command, args []string) error {
			opts, err := resolveMigrationOptions(cmd.Context(), fs, runner, flags, cmd)
			if err != nil {
				return err
			}
			if opts.DSN == "" {
				return fmt.Errorf("database connection string required: specify --db-url flag or DATABASE_URL env var")
			}

			if err := gRunner.Up(cmd.Context(), opts); err != nil {
				return fmt.Errorf("migrating up: %w", err)
			}

			cliOpts := GetOptions(cmd.Context())
			if cliOpts.JSON {
				_ = json.NewEncoder(cmd.OutOrStdout()).Encode(map[string]string{
					"status":  "success",
					"command": "up",
				})
			} else if !cliOpts.Quiet {
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Successfully applied migrations.")
			}
			return nil
		},
	}
}

func newMigrateDownCmd(fs filesystem.FileSystem, runner process.Runner, flags *migrateFlags, gRunner *database.GooseRunner) *cobra.Command {
	var toVersion int64

	cmd := &cobra.Command{
		Use:   "down",
		Short: "Roll back the latest database migration batch (or to specific version)",
		RunE: func(cmd *cobra.Command, args []string) error {
			opts, err := resolveMigrationOptions(cmd.Context(), fs, runner, flags, cmd)
			if err != nil {
				return err
			}
			if opts.DSN == "" {
				return fmt.Errorf("database connection string required: specify --db-url flag or DATABASE_URL env var")
			}

			if cmd.Flags().Changed("to") {
				if toVersion < 0 {
					return fmt.Errorf("invalid rollback version %d: version must be >= 0", toVersion)
				}
				if err := gRunner.DownTo(cmd.Context(), opts, toVersion); err != nil {
					return fmt.Errorf("migrating down to %d: %w", toVersion, err)
				}
			} else {
				if err := gRunner.Down(cmd.Context(), opts); err != nil {
					return fmt.Errorf("migrating down: %w", err)
				}
			}

			cliOpts := GetOptions(cmd.Context())
			if cliOpts.JSON {
				_ = json.NewEncoder(cmd.OutOrStdout()).Encode(map[string]string{
					"status":  "success",
					"command": "down",
				})
			} else if !cliOpts.Quiet {
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Successfully rolled back migration.")
			}
			return nil
		},
	}

	cmd.Flags().Int64Var(&toVersion, "to", 0, "Rollback down to a specific migration version")
	return cmd
}

func newMigrateStatusCmd(fs filesystem.FileSystem, runner process.Runner, flags *migrateFlags, gRunner *database.GooseRunner) *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Dump the status of all migrations",
		RunE: func(cmd *cobra.Command, args []string) error {
			opts, err := resolveMigrationOptions(cmd.Context(), fs, runner, flags, cmd)
			if err != nil {
				return err
			}
			if opts.DSN == "" {
				return fmt.Errorf("database connection string required: specify --db-url flag or DATABASE_URL env var")
			}

			if err := gRunner.Status(cmd.Context(), opts); err != nil {
				return fmt.Errorf("retrieving migration status: %w", err)
			}
			return nil
		},
	}
}

func newMigrateCreateCmd(fs filesystem.FileSystem, runner process.Runner, flags *migrateFlags, gRunner *database.GooseRunner) *cobra.Command {
	return &cobra.Command{
		Use:   "create <name>",
		Short: "Create a new timestamped SQL migration file",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			opts, err := resolveMigrationOptions(cmd.Context(), fs, runner, flags, cmd)
			if err != nil {
				return err
			}

			filePath, err := gRunner.Create(cmd.Context(), opts.Directory, name)
			if err != nil {
				return fmt.Errorf("creating migration %s: %w", name, err)
			}

			cliOpts := GetOptions(cmd.Context())
			if cliOpts.JSON {
				_ = json.NewEncoder(cmd.OutOrStdout()).Encode(map[string]string{
					"status": "created",
					"file":   filePath,
				})
			} else if !cliOpts.Quiet {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Created migration: %s\n", filePath)
			}
			return nil
		},
	}
}

func newMigrateRedoCmd(fs filesystem.FileSystem, runner process.Runner, flags *migrateFlags, gRunner *database.GooseRunner) *cobra.Command {
	return &cobra.Command{
		Use:   "redo",
		Short: "Roll back the most recent migration and re-apply it",
		RunE: func(cmd *cobra.Command, args []string) error {
			opts, err := resolveMigrationOptions(cmd.Context(), fs, runner, flags, cmd)
			if err != nil {
				return err
			}
			if opts.DSN == "" {
				return fmt.Errorf("database connection string required: specify --db-url flag or DATABASE_URL env var")
			}

			if err := gRunner.Redo(cmd.Context(), opts); err != nil {
				return fmt.Errorf("redoing migration: %w", err)
			}

			cliOpts := GetOptions(cmd.Context())
			if cliOpts.JSON {
				_ = json.NewEncoder(cmd.OutOrStdout()).Encode(map[string]string{
					"status":  "success",
					"command": "redo",
				})
			} else if !cliOpts.Quiet {
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Successfully redid migration.")
			}
			return nil
		},
	}
}

func newMigrateResetCmd(fs filesystem.FileSystem, runner process.Runner, flags *migrateFlags, gRunner *database.GooseRunner) *cobra.Command {
	return &cobra.Command{
		Use:   "reset",
		Short: "Roll back all database migrations",
		RunE: func(cmd *cobra.Command, args []string) error {
			opts, err := resolveMigrationOptions(cmd.Context(), fs, runner, flags, cmd)
			if err != nil {
				return err
			}
			if opts.DSN == "" {
				return fmt.Errorf("database connection string required: specify --db-url flag or DATABASE_URL env var")
			}

			if err := gRunner.Reset(cmd.Context(), opts); err != nil {
				return fmt.Errorf("resetting migrations: %w", err)
			}

			cliOpts := GetOptions(cmd.Context())
			if cliOpts.JSON {
				_ = json.NewEncoder(cmd.OutOrStdout()).Encode(map[string]string{
					"status":  "success",
					"command": "reset",
				})
			} else if !cliOpts.Quiet {
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Successfully reset all migrations.")
			}
			return nil
		},
	}
}

func newMigrateVersionCmd(fs filesystem.FileSystem, runner process.Runner, flags *migrateFlags, gRunner *database.GooseRunner) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the current database migration version",
		RunE: func(cmd *cobra.Command, args []string) error {
			opts, err := resolveMigrationOptions(cmd.Context(), fs, runner, flags, cmd)
			if err != nil {
				return err
			}
			if opts.DSN == "" {
				return fmt.Errorf("database connection string required: specify --db-url flag or DATABASE_URL env var")
			}

			ver, err := gRunner.Version(cmd.Context(), opts)
			if err != nil {
				return fmt.Errorf("getting migration version: %w", err)
			}

			cliOpts := GetOptions(cmd.Context())
			if cliOpts.JSON {
				_ = json.NewEncoder(cmd.OutOrStdout()).Encode(map[string]string{
					"version": strconv.FormatInt(ver, 10),
				})
			} else if !cliOpts.Quiet {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Current database version: %d\n", ver)
			}
			return nil
		},
	}
}
