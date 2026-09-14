package cli

import (
	"fmt"
	"path/filepath"

	"github.com/loy-go/loy/internal/diagnostics"
	"github.com/loy-go/loy/internal/filesystem"
	"github.com/loy-go/loy/internal/process"
	"github.com/spf13/cobra"
)

func newDocCmd(fs filesystem.FileSystem, runner process.Runner) *cobra.Command {
	var outputDir string

	cmd := &cobra.Command{
		Use:   "doc [path]",
		Short: "Generate OpenAPI / Swagger documentation from code annotations",
		Long:  `loy doc parses route handlers and doc annotations to generate OpenAPI specifications.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			rootDir := "."
			if len(args) > 0 {
				rootDir = args[0]
			}

			if outputDir == "" {
				outputDir = "docs"
			}

			mainEntry := "cmd/api/main.go"
			webPath := filepath.Join(rootDir, "cmd/web/main.go")
			if exists, _ := fs.Exists(webPath); exists {
				mainEntry = "cmd/web/main.go"
			}

			if runner == nil {
				runner = process.NewExecRunner()
			}

			res, err := runner.Run(cmd.Context(), rootDir, "swag", "init", "-g", mainEntry, "-o", outputDir)
			if err != nil {
				return &CommandError{
					Code: 1,
					Diagnostics: []*diagnostics.Diagnostic{{
						Severity: diagnostics.SeverityError,
						Code:     diagnostics.CodeCLIExecutionFail,
						Message:  fmt.Sprintf("generating OpenAPI documentation with swag: %v", err),
						Detail:   string(res.Stderr),
						Hint:     "install swag via 'go install github.com/swaggo/swag/cmd/swag@latest'",
						File:     rootDir,
					}},
				}
			}

			cliOpts := GetOptions(cmd.Context())
			if !cliOpts.Quiet {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Successfully generated OpenAPI documentation in %s/\n", outputDir)
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&outputDir, "output", "o", "docs", "Output directory for OpenAPI specifications")
	return cmd
}
