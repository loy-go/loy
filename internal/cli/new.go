package cli

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/loy-go/loy/internal/diagnostics"
	"github.com/loy-go/loy/internal/filesystem"
	"github.com/loy-go/loy/internal/generator"
	"github.com/loy-go/loy/internal/generator/builtin"
	"github.com/loy-go/loy/internal/generator/plan"
	"github.com/loy-go/loy/internal/preset"
	"github.com/loy-go/loy/internal/process"
)

func newNewCmd(fs filesystem.FileSystem, runner process.Runner) *cobra.Command {
	var (
		presetName     string
		force          bool
		httpOpt        string
		dbOpt          string
		queueOpt       string
		cacheOpt       string
		multiTenantOpt string
	)

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

			if httpOpt != "" {
				p.Defaults.HTTP = httpOpt
			}
			if dbOpt != "" {
				if dbOpt == "none" {
					p.Defaults.Database = ""
				} else {
					p.Defaults.Database = dbOpt
				}
			}
			if queueOpt != "" {
				if queueOpt == "none" {
					p.Defaults.Queue = ""
				} else {
					p.Defaults.Queue = queueOpt
				}
			}
			if cacheOpt != "" {
				if cacheOpt == "none" {
					p.Defaults.Cache = ""
				} else {
					p.Defaults.Cache = cacheOpt
				}
			}

			if err := fs.MkdirAll(targetDir, 0755); err != nil {
				return fmt.Errorf("creating project directory: %w", err)
			}

			// Initialize go.mod
			goModPath, _ := filesystem.CleanAndValidatePath(targetDir, filepath.Join(targetDir, "go.mod"))
			_ = fs.WriteFile(goModPath, fmt.Appendf(nil, "module %s\n\ngo 1.22\n", projectName), 0644)

			manifestPath, _ := filesystem.CleanAndValidatePath(targetDir, filepath.Join(targetDir, "loy.yaml"))
			yamlContent := p.MaterializeYAML(projectName)
			if multiTenantOpt != "" {
				yamlContent += fmt.Sprintf("\nmulti_tenancy:\n  enabled: true\n  strategy: %s\n", multiTenantOpt)
			}
			if err := fs.WriteFile(manifestPath, []byte(yamlContent), 0644); err != nil {
				return fmt.Errorf("writing loy.yaml: %w", err)
			}

			// Generate initial runtime scaffolding if not a monorepo workspace
			if p.Workspace == nil {
				httpFw := p.Defaults.HTTP
				if httpFw == "" {
					httpFw = "fiber"
				}
				runtimeGen := builtin.NewRuntimeGenerator(projectName, httpFw).
					WithCapabilities(p.Defaults.Database != "", p.Defaults.Cache != "", p.Defaults.Queue != "", true).
					WithTemplate(p.Defaults.Template).
					WithAssets(p.Defaults.Assets)

				if p.Defaults.Template != "" || p.Name == "web" || p.Name == "fullstack" {
					runtimeGen = runtimeGen.WithEntrypoint("web")
				}

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
					_, _ = fmt.Fprintf(cmd.OutOrStdout(), `{"status":"created","project":%q,"directory":%q,"preset":%q}`+"\n", projectName, targetDir, presetName)
				} else {
					_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Created new Loy project %q (%s) in %s\n", projectName, presetName, targetDir)
				}
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&presetName, "preset", "p", "api", "Preset template (api, fullstack, minimal, monorepo, web)")
	cmd.Flags().BoolVarP(&force, "force", "f", false, "Overwrite destination directory if exists")
	cmd.Flags().StringVar(&httpOpt, "http", "", "HTTP adapter (fiber, chi, nethttp, echo)")
	cmd.Flags().StringVar(&dbOpt, "db", "", "Database adapter (postgres, sqlite, mysql, none)")
	cmd.Flags().StringVar(&queueOpt, "queue", "", "Queue adapter (asynq, river, none)")
	cmd.Flags().StringVar(&cacheOpt, "cache", "", "Cache adapter (valkey, redis, memory, none)")
	cmd.Flags().StringVar(&multiTenantOpt, "multi-tenant", "", "Multi-tenancy strategy (rls, column)")

	return cmd
}
