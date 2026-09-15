package cli

import (
	"bufio"
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
		interactive    bool
		httpOpt        string
		dbOpt          string
		queueOpt       string
		cacheOpt       string
		multiTenantOpt string
	)

	cmd := &cobra.Command{
		Use:   "new [project-name]",
		Short: "Create a new Loy project with preset configuration",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var projectName string
			if len(args) > 0 {
				projectName = args[0]
			}
			ctx := cmd.Context()
			opts := GetOptions(ctx)

			if interactive || (len(args) == 0 && !opts.JSON) {
				var err error
				projectName, presetName, httpOpt, dbOpt, cacheOpt, queueOpt, err = runInteractiveWizard(cmd, projectName, presetName, httpOpt, dbOpt, cacheOpt, queueOpt)
				if err != nil {
					return err
				}
			}

			if projectName == "" {
				return &CommandError{
					Code: 2,
					Diagnostics: []*diagnostics.Diagnostic{{
						Severity: diagnostics.SeverityError,
						Code:     diagnostics.CodeConfigValidationError,
						Message:  "missing required project name argument (or use --interactive)",
						Hint:     "run 'loy new <project-name>' or 'loy new --interactive'",
					}},
				}
			}

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
				dbDialect := p.Defaults.Database
				if dbDialect == "none" {
					dbDialect = ""
				}
				runtimeGen := builtin.NewRuntimeGenerator(projectName, httpFw).
					WithDatabaseDialect(dbDialect).
					WithCapabilities(dbDialect != "", p.Defaults.Cache != "" && p.Defaults.Cache != "none", p.Defaults.Queue != "" && p.Defaults.Queue != "none", true).
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
	cmd.Flags().BoolVarP(&interactive, "interactive", "i", false, "Interactive project setup wizard")
	cmd.Flags().StringVar(&httpOpt, "http", "", "HTTP adapter (fiber, chi, gin, nethttp, echo)")
	cmd.Flags().StringVar(&dbOpt, "db", "", "Database adapter (postgres, sqlite, mysql, none)")
	cmd.Flags().StringVar(&queueOpt, "queue", "", "Queue adapter (asynq, river, none)")
	cmd.Flags().StringVar(&cacheOpt, "cache", "", "Cache adapter (valkey, redis, memory, none)")
	cmd.Flags().StringVar(&multiTenantOpt, "multi-tenant", "", "Multi-tenancy strategy (rls, column)")

	return cmd
}

func runInteractiveWizard(cmd *cobra.Command, initialName, defaultPreset, defaultHTTP, defaultDB, defaultCache, defaultQueue string) (name, presetName, httpOpt, dbOpt, cacheOpt, queueOpt string, err error) {
	reader := bufio.NewReader(cmd.InOrStdin())
	out := cmd.OutOrStdout()

	_, _ = fmt.Fprintln(out, "🚀 Welcome to Loy Project Setup Wizard")
	_, _ = fmt.Fprintln(out, "Press Enter to accept defaults shown in brackets.")
	_, _ = fmt.Fprintln(out, "")

	// 1. Project Name
	for {
		prompt := "Project name"
		if initialName != "" {
			prompt += fmt.Sprintf(" [%s]", initialName)
		}
		prompt += ": "
		_, _ = fmt.Fprint(out, prompt)

		line, readErr := reader.ReadString('\n')
		if readErr != nil && len(line) == 0 {
			if initialName != "" {
				name = initialName
				break
			}
			return "", "", "", "", "", "", fmt.Errorf("reading project name: %w", readErr)
		}
		line = strings.TrimSpace(line)
		if line == "" && initialName != "" {
			name = initialName
			break
		} else if line != "" {
			name = line
			break
		}
		_, _ = fmt.Fprintln(out, "Project name cannot be empty.")
	}

	// 2. Preset
	if defaultPreset == "" {
		defaultPreset = "api"
	}
	_, _ = fmt.Fprintf(out, "Preset [api, fullstack, minimal, web, monorepo] [%s]: ", defaultPreset)
	if line, err := reader.ReadString('\n'); err == nil {
		line = strings.TrimSpace(line)
		if line != "" {
			presetName = line
		} else {
			presetName = defaultPreset
		}
	} else {
		presetName = defaultPreset
	}

	// 3. HTTP Transport
	if defaultHTTP == "" {
		defaultHTTP = "fiber"
	}
	_, _ = fmt.Fprintf(out, "HTTP Transport [fiber, chi, gin, nethttp] [%s]: ", defaultHTTP)
	if line, err := reader.ReadString('\n'); err == nil {
		line = strings.TrimSpace(line)
		if line != "" {
			httpOpt = line
		} else {
			httpOpt = defaultHTTP
		}
	} else {
		httpOpt = defaultHTTP
	}

	// 4. Database
	if defaultDB == "" {
		defaultDB = "postgres"
	}
	_, _ = fmt.Fprintf(out, "Database [postgres, sqlite, mysql, none] [%s]: ", defaultDB)
	if line, err := reader.ReadString('\n'); err == nil {
		line = strings.TrimSpace(line)
		if line != "" {
			dbOpt = line
		} else {
			dbOpt = defaultDB
		}
	} else {
		dbOpt = defaultDB
	}

	// 5. Cache
	if defaultCache == "" {
		defaultCache = "valkey"
	}
	_, _ = fmt.Fprintf(out, "Cache [valkey, redis, none] [%s]: ", defaultCache)
	if line, err := reader.ReadString('\n'); err == nil {
		line = strings.TrimSpace(line)
		if line != "" {
			cacheOpt = line
		} else {
			cacheOpt = defaultCache
		}
	} else {
		cacheOpt = defaultCache
	}

	// 6. Queue
	if defaultQueue == "" {
		defaultQueue = "asynq"
	}
	_, _ = fmt.Fprintf(out, "Background Queue [asynq, none] [%s]: ", defaultQueue)
	if line, err := reader.ReadString('\n'); err == nil {
		line = strings.TrimSpace(line)
		if line != "" {
			queueOpt = line
		} else {
			queueOpt = defaultQueue
		}
	} else {
		queueOpt = defaultQueue
	}

	_, _ = fmt.Fprintln(out, "")
	return name, presetName, httpOpt, dbOpt, cacheOpt, queueOpt, nil
}
