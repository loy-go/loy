package cli

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/loy-go/loy/internal/diagnostics"
	"github.com/loy-go/loy/internal/filesystem"
	"github.com/loy-go/loy/internal/process"
	"github.com/loy-go/loy/internal/updater"
)

func newSelfUpdateCmd(fs filesystem.FileSystem, runner process.Runner) *cobra.Command {
	var (
		checkOnly bool
		targetVer string
	)

	cmd := &cobra.Command{
		Use:     "self-update",
		Aliases: []string{"update"},
		Short:   "Update the Loy CLI binary to the latest or specified release",
		Long: `loy self-update checks GitHub Releases for the latest verified binary matching your operating system and architecture,
validates the SHA256 cryptographic checksum against release signatures, and atomically replaces the running executable.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			opts := GetOptions(ctx)

			up, err := updater.NewUpdater("", nil)
			if err != nil {
				return &CommandError{
					Code: 1,
					Diagnostics: []*diagnostics.Diagnostic{{
						Severity: diagnostics.SeverityError,
						Code:     diagnostics.CodeInternalError,
						Message:  fmt.Sprintf("initializing updater: %v", err),
					}},
				}
			}

			if checkOnly {
				checkRes, info, err := up.Check(ctx, targetVer)
				if err != nil {
					return &CommandError{
						Code: 1,
						Diagnostics: []*diagnostics.Diagnostic{{
							Severity: diagnostics.SeverityError,
							Code:     diagnostics.CodeCLIExecutionFail,
							Message:  fmt.Sprintf("checking release: %v", err),
						}},
					}
				}

				if opts.JSON {
					return json.NewEncoder(cmd.OutOrStdout()).Encode(checkRes)
				}

				if !opts.Quiet {
					if checkRes.UpdateAvailable {
						_, _ = fmt.Fprintf(cmd.OutOrStdout(), "A new version of Loy is available: %s (current: %s)\n", info.TagName, checkRes.CurrentVersion)
						if checkRes.IsManagedPackage {
							_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Note: Loy is managed via %s. Update with: brew upgrade loy\n", checkRes.PackageManager)
						} else {
							_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Run 'loy self-update' to apply the update.")
						}
					} else {
						_, _ = fmt.Fprintf(cmd.OutOrStdout(), "You are already on the latest version of Loy (%s).\n", checkRes.CurrentVersion)
					}
				}
				return nil
			}

			// Apply update
			if !opts.Quiet && !opts.JSON {
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Checking and applying Loy update...")
			}

			applyRes, err := up.Apply(ctx, targetVer)
			if err != nil {
				return &CommandError{
					Code: 1,
					Diagnostics: []*diagnostics.Diagnostic{{
						Severity: diagnostics.SeverityError,
						Code:     diagnostics.CodeCLIExecutionFail,
						Message:  err.Error(),
					}},
				}
			}

			if opts.JSON {
				return json.NewEncoder(cmd.OutOrStdout()).Encode(applyRes)
			}

			if !opts.Quiet {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Successfully updated Loy from %s to %s\nBinary location: %s\n",
					applyRes.PreviousVersion, applyRes.UpdatedVersion, applyRes.ExecutablePath)
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&checkOnly, "check-only", false, "Check if an update is available without downloading or applying")
	cmd.Flags().StringVar(&targetVer, "version", "", "Target specific release version tag (e.g. v0.3.0)")

	return cmd
}
