package cli

import (
	"encoding/json"
	"fmt"

	"github.com/loy-go/loy/internal/diagnostics"
	"github.com/loy-go/loy/internal/filesystem"
	"github.com/loy-go/loy/internal/process"
	"github.com/loy-go/loy/internal/typegen/typescript"
	"github.com/spf13/cobra"
)

func newGenCmd(fs filesystem.FileSystem, runner process.Runner) *cobra.Command {
	genCmd := &cobra.Command{
		Use:     "gen",
		Aliases: []string{"generate"},
		Short:   "Generate client SDKs, types, and external schemas from application code",
		Long:    `loy gen inspects transport layer definitions to generate strongly-typed API clients and definitions.`,
	}

	genCmd.AddCommand(newGenClientCmd(fs, runner))
	return genCmd
}

func newGenClientCmd(fs filesystem.FileSystem, runner process.Runner) *cobra.Command {
	var outDir string

	cmd := &cobra.Command{
		Use:     "client [root-dir]",
		Short:   "Generate zero-dependency TypeScript types and API client from Go transport DTOs",
		Long:    `loy gen client statically analyzes Go request/resource DTOs and registered routes to generate types.ts and client.ts.`,
		Args:    cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			rootDir := "."
			if len(args) > 0 {
				rootDir = args[0]
			}

			cliOpts := GetOptions(cmd.Context())

			result, err := typescript.Generate(fs, typescript.GenerateOptions{
				RootDir: rootDir,
				OutDir:  outDir,
			})
			if err != nil {
				return &CommandError{
					Code: 1,
					Diagnostics: []*diagnostics.Diagnostic{{
						Severity: diagnostics.SeverityError,
						Code:     diagnostics.CodeGenExecutionError,
						Message:  fmt.Sprintf("generating typescript client: %v", err),
					}},
				}
			}

			if cliOpts.JSON {
				out := map[string]string{
					"types_path":  result.TypesPath,
					"client_path": result.ClientPath,
				}
				return json.NewEncoder(cmd.OutOrStdout()).Encode(out)
			}

			if !cliOpts.Quiet {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Successfully generated TypeScript client in %s\n", outDir)
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "  + %s\n", result.TypesPath)
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "  + %s\n", result.ClientPath)
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&outDir, "out", "o", "client", "Output directory for generated TypeScript files")

	return cmd
}
