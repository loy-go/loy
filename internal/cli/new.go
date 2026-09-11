package cli

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/uloydev/loy/internal/diagnostics"
	"github.com/uloydev/loy/internal/filesystem"
	"github.com/uloydev/loy/internal/generator"
	"github.com/uloydev/loy/internal/generator/builtin"
	"github.com/uloydev/loy/internal/generator/plan"
	"github.com/uloydev/loy/internal/preset"
	"github.com/uloydev/loy/internal/process"
)

func newNewCmd(fs filesystem.FileSystem, runner process.Runner) *cobra.Command {
	var presetName string
	var force bool

	cmd := &cobra.Command{
		Use:   "new <project-name>",
		Short: "Create a new Loy project with preset configuration",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			projectName := args[0]
			ctx := cmd.Context()
			opts := GetOptions(ctx)

			baseDir := "."

			cleanBase, pathErr := filesystem.CleanAndValidatePath(baseDir, baseDir)
			if pathErr != nil {
				return &CommandError{
					Code: 1,
					Diagnostics: []*diagnostics.Diagnostic{{
						Severity: diagnostics.SeverityError,
						Code:     diagnostics.CodeFSPathTraversal,
						Message:  fmt.Sprintf("validating base directory: %v", pathErr),
						File:     baseDir,
					}},
				}
			}

			targetDir, pathErr := filesystem.CleanAndValidatePath(cleanBase, filepath.Join(cleanBase, projectName))
			if pathErr != nil {
				return &CommandError{
					Code: 1,
					Diagnostics: []*diagnostics.Diagnostic{{
						Severity: diagnostics.SeverityError,
						Code:     diagnostics.CodeFSPathTraversal,
						Message:  fmt.Sprintf("invalid project name escapes directory: %v", pathErr),
						File:     projectName,
					}},
				}
			}

			exists, _ := fs.Exists(targetDir)
			if exists && !force {
				return &CommandError{
					Code: 1,
					Diagnostics: []*diagnostics.Diagnostic{{
						Severity: diagnostics.SeverityError,
						Code:     diagnostics.CodeProjectAlreadyExists,
						Message:  fmt.Sprintf("directory %q already exists", targetDir),
						Hint:     "use --force to overwrite or choose a different name",
						File:     targetDir,
					}},
				}
			}

			reg := preset.NewRegistry()
			p, ok := reg.Get(presetName)
			if !ok {
				return &CommandError{
					Code: 2,
					Diagnostics: []*diagnostics.Diagnostic{{
						Severity: diagnostics.SeverityError,
						Code:     diagnostics.CodeConfigPresetNotFound,
						Message:  fmt.Sprintf("unknown preset %q", presetName),
						Hint:     fmt.Sprintf("available presets: %s", strings.Join(reg.Names(), ", ")),
						File:     targetDir,
					}},
				}
			}

			if err := fs.MkdirAll(targetDir, 0755); err != nil {
				return fmt.Errorf("creating project directory: %w", err)
			}

			// Initialize go.mod
			goModPath, _ := filesystem.CleanAndValidatePath(targetDir, filepath.Join(targetDir, "go.mod"))
			_ = fs.WriteFile(goModPath, []byte(fmt.Sprintf("module %s\n\ngo 1.22\n", projectName)), 0644)

			manifestPath, _ := filesystem.CleanAndValidatePath(targetDir, filepath.Join(targetDir, "loy.yaml"))
			yamlContent := p.MaterializeYAML(projectName)
			if err := fs.WriteFile(manifestPath, []byte(yamlContent), 0644); err != nil {
				return fmt.Errorf("writing loy.yaml: %w", err)
			}

			// Generate initial runtime scaffolding if not a monorepo workspace
			if p.Workspace == nil {
				httpFw := p.Defaults.HTTP
				if httpFw == "" {
					httpFw = "fiber"
				}
				runtimeGen := builtin.NewRuntimeGenerator(projectName, httpFw)
				artifacts, err := runtimeGen.Generate(ctx, generator.Input{Name: "runtime"})
				if err != nil {
					return fmt.Errorf("scaffolding runtime: %w", err)
				}
				builder := plan.NewBuilder(fs)
				runtimePlan, err := builder.Build(ctx, targetDir, artifacts, generator.Options{Force: force})
				if err != nil {
					return fmt.Errorf("building runtime plan: %w", err)
				}
				executor := plan.NewExecutor(fs)
				if err := executor.Execute(ctx, runtimePlan); err != nil {
					return fmt.Errorf("executing runtime plan: %w", err)
				}
			}

			if runner != nil {
				_, _ = runner.Run(ctx, targetDir, "go", "mod", "tidy")
			}

			if !opts.Quiet {
				if opts.JSON {
					fmt.Printf(`{"status":"created","project":%q,"directory":%q,"preset":%q}`+"\n", projectName, targetDir, presetName)
				} else {
					fmt.Printf("Created new Loy project %q (%s) in %s\n", projectName, presetName, targetDir)
				}
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&presetName, "preset", "p", "api", "Preset template (api, fullstack, minimal, monorepo, web)")
	cmd.Flags().BoolVarP(&force, "force", "f", false, "Overwrite destination directory if exists")

	return cmd
}
