package cli

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/loy-go/loy/internal/diagnostics"
	"github.com/loy-go/loy/internal/doctor"
	"github.com/loy-go/loy/internal/filesystem"
	"github.com/loy-go/loy/internal/process"
)

func newDoctorCmd(fs filesystem.FileSystem, runner process.Runner) *cobra.Command {
	var strict bool

	cmd := &cobra.Command{
		Use:   "doctor [path]",
		Short: "Validate development environment, toolchain prerequisites, and project state",
		Long:  `loy doctor inspects host tools (Go, Git, Docker, sqlc) and project validity without modifying the system.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			targetDir := "."
			if len(args) > 0 {
				targetDir = args[0]
			}

			cliOpts := GetOptions(cmd.Context())

			doc := doctor.NewDoctor(fs, runner)
			checks, err := doc.Run(cmd.Context(), targetDir)
			if err != nil {
				return fmt.Errorf("running doctor: %w", err)
			}

			hasErrors := false
			hasWarnings := false
			var diags []diagnostics.Diagnostic

			for _, c := range checks {
				if !c.Passed {
					if c.Diagnostic != nil {
						diags = append(diags, *c.Diagnostic)
						if c.Diagnostic.Severity == diagnostics.SeverityError {
							hasErrors = true
						} else {
							hasWarnings = true
						}
					} else {
						hasErrors = true
					}
				}
			}

			if cliOpts.JSON {
				if err := json.NewEncoder(cmd.OutOrStdout()).Encode(checks); err != nil {
					return err
				}
			} else {
				for _, c := range checks {
					if c.Passed {
						_, _ = fmt.Fprintf(cmd.OutOrStdout(), "[✓] %-15s : %s\n", c.Name, c.Detail)
					} else {
						tag := "[!]"
						if c.Diagnostic != nil && c.Diagnostic.Severity == diagnostics.SeverityError {
							tag = "[x]"
						}
						_, _ = fmt.Fprintf(cmd.OutOrStdout(), "%s %-15s : %s\n", tag, c.Name, c.Detail)
						if c.Diagnostic != nil && c.Diagnostic.Hint != "" {
							_, _ = fmt.Fprintf(cmd.OutOrStdout(), "    Hint: %s\n", c.Diagnostic.Hint)
						}
					}
				}
			}

			if hasErrors || (strict && hasWarnings) {
				var errDiags []*diagnostics.Diagnostic
				for i := range diags {
					errDiags = append(errDiags, &diags[i])
				}
				return &CommandError{
					Code:        1,
					Diagnostics: errDiags,
				}
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&strict, "strict", false, "Treat warnings as errors and exit with code 1")

	return cmd
}
