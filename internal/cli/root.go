package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/loy-go/loy/internal/filesystem"
	"github.com/loy-go/loy/internal/process"
	"github.com/loy-go/loy/internal/version"
)

type contextKey string

const optionsKey contextKey = "loy.cli.options"

// GlobalOptions contains the global CLI flags.
type GlobalOptions struct {
	JSON      bool
	Verbose   bool
	Quiet     bool
	NoColor   bool
	Directory string
}

// WithOptions stores GlobalOptions in a Context.
func WithOptions(ctx context.Context, opts *GlobalOptions) context.Context {
	return context.WithValue(ctx, optionsKey, opts)
}

// GetOptions retrieves GlobalOptions from a Context.
func GetOptions(ctx context.Context) *GlobalOptions {
	if opts, ok := ctx.Value(optionsKey).(*GlobalOptions); ok && opts != nil {
		return opts
	}
	return &GlobalOptions{}
}

// NewRootCmdWithFS initializes the root cobra command with custom filesystem and runner.
func NewRootCmdWithFS(fs filesystem.FileSystem, runner process.Runner) *cobra.Command {
	opts := &GlobalOptions{}

	rootCmd := &cobra.Command{
		Use:           "loy",
		Short:         "Loy — Go developer platform with Laravel-like DX",
		Long:          `Loy is an opinionated developer platform and CLI toolchain for Go, providing architecture enforcement, typed code generation, and runtime conventions.`,
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		cmd.SetContext(WithOptions(cmd.Context(), opts))
		if opts.Directory != "" {
			if err := os.Chdir(opts.Directory); err != nil {
				return fmt.Errorf("changing working directory to %s: %w", opts.Directory, err)
			}
		}
		return nil
	}

	rootCmd.SetContext(WithOptions(context.Background(), opts))

	rootCmd.PersistentFlags().BoolVar(&opts.JSON, "json", false, "Output results in JSON format")
	rootCmd.PersistentFlags().BoolVarP(&opts.Verbose, "verbose", "v", false, "Enable verbose/debug output")
	rootCmd.PersistentFlags().BoolVarP(&opts.Quiet, "quiet", "q", false, "Suppress non-essential output")
	rootCmd.PersistentFlags().BoolVar(&opts.NoColor, "no-color", false, "Disable colored ANSI output")
	rootCmd.PersistentFlags().StringVarP(&opts.Directory, "directory", "C", "", "Change execution directory")

	if fs == nil {
		fs = filesystem.NewOSFileSystem()
	}
	if runner == nil {
		runner = process.NewExecRunner()
	}

	rootCmd.AddCommand(newVersionCmd())
	rootCmd.AddCommand(newInitCmd(fs, runner))
	rootCmd.AddCommand(newNewCmd(fs, runner))
	rootCmd.AddCommand(newCheckCmd(fs, runner))
	rootCmd.AddCommand(newMakeCmd(fs, runner))
	rootCmd.AddCommand(newMigrateCmd(fs, runner))
	rootCmd.AddCommand(newDevCmd(fs, runner))
	rootCmd.AddCommand(newDoctorCmd(fs, runner))
	rootCmd.AddCommand(newGraphCmd(fs, runner))
	rootCmd.AddCommand(newHookCmd(fs, runner))
	rootCmd.AddCommand(newDocCmd(fs, runner))
	rootCmd.AddCommand(newRoutesCmd(fs, runner))
	rootCmd.AddCommand(newSeedCmd(fs, runner))
	rootCmd.AddCommand(newUpgradeCmd(fs, runner))
	rootCmd.AddCommand(newCompletionCmd())

	return rootCmd
}

// NewRootCmd initializes and configures the root cobra command using OS defaults.
func NewRootCmd() *cobra.Command {
	return NewRootCmdWithFS(nil, nil)
}

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the Loy CLI version",
		RunE: func(cmd *cobra.Command, args []string) error {
			opts := GetOptions(cmd.Context())
			info := version.Get()
			if opts.JSON {
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(info)
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "loy version %s (commit: %s, date: %s, %s)\n", info.Version, info.Commit, info.BuildDate, info.Platform)
			return nil
		},
	}
}
