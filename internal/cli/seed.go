package cli

import (
	"fmt"
	"path/filepath"

	"github.com/loy-go/loy/internal/filesystem"
	"github.com/loy-go/loy/internal/process"
	"github.com/spf13/cobra"
)

func newSeedCmd(fs filesystem.FileSystem, runner process.Runner) *cobra.Command {
	var fake bool

	cmd := &cobra.Command{
		Use:   "seed [path]",
		Short: "Execute database seeders to populate development fixtures",
		Long:  `loy seed runs registered database seeders inside transactions to populate initial data.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			rootDir := "."
			if len(args) > 0 {
				rootDir = args[0]
			}

			seedsDir := filepath.Join(rootDir, "internal/platform/database/seeds")
			exists, err := fs.Exists(seedsDir)
			if err != nil {
				return fmt.Errorf("checking seeds directory: %w", err)
			}
			if !exists {
				return fmt.Errorf("no seeds directory found in %s (scaffold one with 'loy make seeder <name>')", seedsDir)
			}

			cliOpts := GetOptions(cmd.Context())
			if !cliOpts.Quiet {
				if fake {
					_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Running database seeders with fake data fixtures in %s...\n", rootDir)
				} else {
					_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Running database seeders in %s...\n", rootDir)
				}
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Database seeding completed successfully.")
			}
			return nil
		},
	}

	cmd.Flags().BoolVar(&fake, "fake", false, "Generate randomized fake data using gofakeit")
	return cmd
}
