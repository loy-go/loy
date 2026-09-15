package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/loy-go/loy/internal/diagnostics"
	"github.com/loy-go/loy/internal/discovery"
	"github.com/loy-go/loy/internal/filesystem"
	"github.com/loy-go/loy/internal/generator"
	"github.com/loy-go/loy/internal/generator/builtin"
	"github.com/loy-go/loy/internal/generator/builtin/wiring"
	"github.com/loy-go/loy/internal/generator/model"
	"github.com/loy-go/loy/internal/generator/plan"
	"github.com/loy-go/loy/internal/manifest"
	"github.com/loy-go/loy/internal/process"
	"github.com/loy-go/loy/internal/workspace"
	"golang.org/x/mod/modfile"
)

type makeOptions struct {
	target  string
	force   bool
	dryRun  bool
	modular bool
}

func newMakeCmd(fs filesystem.FileSystem, runner process.Runner) *cobra.Command {
	opts := &makeOptions{}

	cmd := &cobra.Command{
		Use:   "make <artifact> <name> [flags] [args...]",
		Short: "Scaffold application components, slices, and vertical features",
		Long: `The make command tree provides scaffolding for clean-architecture Go applications:
  - Atomic generators: model, repository (repo), service (svc), handler, request, resource, job, event, listener, policy, test
  - Composers: feature, crud`,
	}

	cmd.PersistentFlags().StringVar(&opts.target, "target", "", "Target application module in workspace")
	cmd.PersistentFlags().BoolVar(&opts.force, "force", false, "Overwrite existing files if developer owned")
	cmd.PersistentFlags().BoolVar(&opts.dryRun, "dry-run", false, "Preview generated operations without writing to disk")
	cmd.PersistentFlags().BoolVar(&opts.modular, "modular", false, "Scaffold sub-domain modular wiring file instead of flat wiring")

	// Atomic generators
	cmd.AddCommand(newArtifactCmd("model", "Scaffold domain entity model", fs, runner, opts, func(mod string) generator.Generator {
		return builtin.NewModelGenerator(mod)
	}, []string{}))

	cmd.AddCommand(newArtifactCmd("repository", "Scaffold domain repository interface & adapter", fs, runner, opts, func(mod string) generator.Generator {
		return builtin.NewRepositoryGenerator(mod)
	}, []string{"repo"}))

	cmd.AddCommand(newArtifactCmd("service", "Scaffold application service use case", fs, runner, opts, func(mod string) generator.Generator {
		return builtin.NewServiceGenerator(mod)
	}, []string{"svc"}))

	cmd.AddCommand(newArtifactCmd("handler", "Scaffold HTTP transport handler", fs, runner, opts, func(mod string) generator.Generator {
		return builtin.NewHandlerGenerator(mod)
	}, []string{}))

	cmd.AddCommand(newArtifactCmd("request", "Scaffold HTTP request DTO with validation", fs, runner, opts, func(mod string) generator.Generator {
		return builtin.NewRequestGenerator(mod)
	}, []string{}))

	cmd.AddCommand(newArtifactCmd("resource", "Scaffold API response resource transformation", fs, runner, opts, func(mod string) generator.Generator {
		return builtin.NewResourceGenerator(mod)
	}, []string{}))

	cmd.AddCommand(newArtifactCmd("job", "Scaffold Asynq background job payload & processor", fs, runner, opts, func(mod string) generator.Generator {
		return builtin.NewJobGenerator(mod)
	}, []string{}))

	cmd.AddCommand(newArtifactCmd("event", "Scaffold domain event struct", fs, runner, opts, func(mod string) generator.Generator {
		return builtin.NewEventGenerator(mod)
	}, []string{}))

	cmd.AddCommand(newArtifactCmd("listener", "Scaffold event listener consumer", fs, runner, opts, func(mod string) generator.Generator {
		return builtin.NewListenerGenerator(mod)
	}, []string{}))

	cmd.AddCommand(newArtifactCmd("policy", "Scaffold authorization policy checks", fs, runner, opts, func(mod string) generator.Generator {
		return builtin.NewPolicyGenerator(mod)
	}, []string{}))

	cmd.AddCommand(newArtifactCmd("auth", "Scaffold baseline JWT and password authentication kit", fs, runner, opts, func(mod string) generator.Generator {
		return builtin.NewAuthGenerator(mod)
	}, []string{}))

	cmd.AddCommand(newArtifactCmd("tenant", "Scaffold multi-tenancy context, RLS helper, and initial migration", fs, runner, opts, func(mod string) generator.Generator {
		return builtin.NewTenantGenerator(mod)
	}, []string{}))

	cmd.AddCommand(newArtifactCmd("grpc", "Scaffold Proto contract and gRPC transport server", fs, runner, opts, func(mod string) generator.Generator {
		return builtin.NewGRPCGenerator(mod)
	}, []string{}))

	cmd.AddCommand(newArtifactCmd("ws", "Scaffold WebSocket hub, client pumps, and protocol frames", fs, runner, opts, func(mod string) generator.Generator {
		return builtin.NewWebSocketGenerator(mod)
	}, []string{"websocket"}))

	cmd.AddCommand(newArtifactCmd("outbox", "Scaffold Transactional Outbox migration, store, and dispatcher", fs, runner, opts, func(mod string) generator.Generator {
		return builtin.NewOutboxGenerator(mod)
	}, []string{}))

	cmd.AddCommand(newArtifactCmd("seeder", "Scaffold database seeder fixture", fs, runner, opts, func(mod string) generator.Generator {
		return builtin.NewSeederGenerator(mod)
	}, []string{"seed"}))

	cmd.AddCommand(newArtifactCmd("test", "Scaffold unit and integration tests", fs, runner, opts, func(mod string) generator.Generator {
		return builtin.NewTestGenerator(mod)
	}, []string{}))

	cmd.AddCommand(newViewCmd(fs, runner, opts))

	cmd.AddCommand(newRuntimeCmd(fs, runner, opts))

	// Deployment & infrastructure generators
	cmd.AddCommand(newDockerCmd(fs, runner, opts))
	cmd.AddCommand(newK8sCmd(fs, runner, opts))
	cmd.AddCommand(newHelmCmd(fs, runner, opts))
	cmd.AddCommand(newCICmd(fs, runner, opts))

	// Composers
	cmd.AddCommand(newFeatureCmd(fs, runner, opts))
	cmd.AddCommand(newCRUDCmd(fs, runner, opts))
	cmd.AddCommand(newDeployCmd(fs, runner, opts))

	return cmd
}

func newArtifactCmd(name, short string, fs filesystem.FileSystem, runner process.Runner, opts *makeOptions, factory func(modulePath string) generator.Generator, aliases []string) *cobra.Command {
	cmd := &cobra.Command{
		Use:     fmt.Sprintf("%s <name> [fields...]", name),
		Short:   short,
		Aliases: aliases,
		Args:    cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			artifactName := args[0]
			extraArgs := strings.Join(args[1:], " ")

			return runGenerator(cmd, fs, runner, opts, func(mod string) generator.Generator {
				return factory(mod)
			}, artifactName, extraArgs, false)
		},
	}
	return cmd
}

func newRuntimeCmd(fs filesystem.FileSystem, runner process.Runner, opts *makeOptions) *cobra.Command {
	var (
		httpFramework string
		dbDialect     string
	)
	cmd := &cobra.Command{
		Use:   "runtime",
		Short: "Scaffold application runtime lifecycle, composition root, and health checks",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runGenerator(cmd, fs, runner, opts, func(mod string) generator.Generator {
				fw := httpFramework
				if fw == "" {
					fw = "fiber"
				}
				gen := builtin.NewRuntimeGenerator(mod, fw)
				if dbDialect != "" {
					gen.WithDatabaseDialect(dbDialect)
				}
				return gen
			}, "runtime", "", false)
		},
	}
	cmd.Flags().StringVar(&httpFramework, "http", "", "HTTP framework to scaffold (fiber, chi, gin, or nethttp)")
	cmd.Flags().StringVar(&dbDialect, "db", "", "Database adapter to scaffold (postgres, sqlite, mysql, or none)")
	return cmd
}

func newFeatureCmd(fs filesystem.FileSystem, runner process.Runner, opts *makeOptions) *cobra.Command {
	return &cobra.Command{
		Use:   "feature <name> [fields...]",
		Short: "Scaffold full vertical feature slice (model, repo, service, handler, test, wiring)",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			featureName := args[0]
			extraArgs := strings.Join(args[1:], " ")

			return runGenerator(cmd, fs, runner, opts, func(mod string) generator.Generator {
				return builtin.NewFeatureGenerator(mod)
			}, featureName, extraArgs, true)
		},
	}
}

func newCRUDCmd(fs filesystem.FileSystem, runner process.Runner, opts *makeOptions) *cobra.Command {
	return &cobra.Command{
		Use:   "crud <name> [fields...]",
		Short: "Scaffold full vertical CRUD slice (migration, queries, model, repo, service, handler, test, wiring)",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			crudName := args[0]
			extraArgs := strings.Join(args[1:], " ")

			err := runGenerator(cmd, fs, runner, opts, func(mod string) generator.Generator {
				return builtin.NewCRUDGenerator(mod)
			}, crudName, extraArgs, true)
			if err != nil {
				return err
			}

			// Post-generation: check if sqlc is available
			if !opts.dryRun {
				checkSQLC(cmd, fs, runner, opts)
			}
			return nil
		},
	}
}

func resolveProjectTarget(ctx context.Context, fs filesystem.FileSystem, runner process.Runner, requestedTarget string) (targetDir string, modulePath string, err error) {
	currDir, _ := os.Getwd()
	if currDir == "" {
		currDir = "."
	}
	disc, err := discovery.NewDiscoverer(fs, runner)
	if err != nil {
		return "", "", err
	}

	discovered, _ := disc.Discover(ctx, currDir)
	if discovered == nil || discovered.RootDir == "" {
		return "", "", &CommandError{
			Code: 1,
			Diagnostics: []*diagnostics.Diagnostic{{
				Severity: diagnostics.SeverityError,
				Code:     diagnostics.CodeProjectRootNotFound,
				Message:  "not inside a Go module or Loy workspace",
				Hint:     "run 'go mod init' or 'loy init' first",
				File:     currDir,
			}},
		}
	}

	targetDir = discovered.RootDir

	// Read full module path from go.mod if present
	if discovered.GoModPath != "" {
		if data, err := fs.ReadFile(discovered.GoModPath); err == nil {
			f, err := modfile.Parse(discovered.GoModPath, data, nil)
			if err == nil && f.Module != nil {
				modulePath = f.Module.Mod.Path
			}
		}
	}

	// Fallback to inferred project name
	if modulePath == "" {
		if inferred, infErr := disc.InferProjectName(targetDir); infErr == nil && inferred != "" {
			modulePath = inferred
		}
	}

	// If workspace present, resolve target
	if discovered.IsWorkspace || discovered.HasGoWork {
		wsResolver, wsErr := workspace.NewResolver(fs)
		if wsErr == nil && wsResolver != nil {
			ws, _ := wsResolver.Resolve(discovered.RootDir)
			if ws != nil {
				mod, diag := ws.SelectTarget(requestedTarget)
				if diag != nil {
					return "", "", &CommandError{Code: 1, Diagnostics: []*diagnostics.Diagnostic{diag}}
				}
				if mod != nil {
					targetDir = mod.Path
					modulePath = mod.Name
				}
			}
		}
	}

	if modulePath == "" {
		modulePath = filepath.Base(targetDir)
	}

	return targetDir, modulePath, nil
}

func runGenerator(cmd *cobra.Command, fs filesystem.FileSystem, runner process.Runner, opts *makeOptions, factory func(modulePath string) generator.Generator, name string, fields string, requiresWiring bool) error {
	ctx := cmd.Context()
	globalOpts := GetOptions(ctx)

	targetDir, modulePath, err := resolveProjectTarget(ctx, fs, runner, opts.target)
	if err != nil {
		return err
	}

	gen := factory(modulePath)
	input := generator.Input{
		Name: name,
		Args: map[string]string{
			"fields":  fields,
			"modular": fmt.Sprintf("%t", opts.modular),
		},
		Options: generator.Options{
			Force:  opts.force,
			DryRun: opts.dryRun,
		},
	}

	// Inspect loy.yaml in targetDir to populate manifest defaults (e.g. multi-tenancy, http, database)
	manifestPath := filepath.Join(targetDir, "loy.yaml")
	if mData, err := fs.ReadFile(manifestPath); err == nil {
		p := manifest.NewParser()
		if m, diag := p.ParseStrict(manifestPath, mData); diag == nil && m != nil {
			if m.MultiTenancy.Enabled {
				input.Args["multi_tenant"] = "true"
				strategy := m.MultiTenancy.Strategy
				if strategy == "" {
					strategy = "rls"
				}
				input.Args["tenant_strategy"] = strategy
			}
			if m.Defaults.HTTP != "" && input.Args["http"] == "" {
				input.Args["http"] = m.Defaults.HTTP
			}
			if m.Defaults.Database != "" && input.Args["database"] == "" {
				input.Args["database"] = m.Defaults.Database
			}
		}
	}

	if httpFlag := cmd.Flags().Lookup("http"); httpFlag != nil && httpFlag.Changed {
		input.Args["http"] = httpFlag.Value.String()
	}
	if dbFlag := cmd.Flags().Lookup("db"); dbFlag != nil && dbFlag.Changed {
		input.Args["database"] = dbFlag.Value.String()
	}

	artifacts, err := gen.Generate(ctx, input)
	if err != nil {
		return &CommandError{
			Code: 1,
			Diagnostics: []*diagnostics.Diagnostic{{
				Severity: diagnostics.SeverityError,
				Code:     diagnostics.CodeGenExecutionError,
				Message:  fmt.Sprintf("generating %s %s: %v", gen.Name(), name, err),
				File:     targetDir,
			}},
		}
	}

	// If feature requires wiring and wiring.go does not exist, include initial template artifact in plan
	if requiresWiring {
		wiringRel := "internal/app/wiring.go"
		wiringFull, pErr := plan.ResolveTargetPath(targetDir, wiringRel)
		if pErr == nil {
			exists, _ := fs.Exists(wiringFull)
			if !exists {
				initArtifact := model.Artifact{
					Path:        wiringRel,
					Content:     []byte(wiring.DefaultWiringTemplate),
					Ownership:   model.DeveloperOwned,
					Permissions: 0644,
				}
				artifacts = append([]model.Artifact{initArtifact}, artifacts...)
			}
		}
	}

	planBuilder := plan.NewBuilder(fs)
	executionPlan, err := planBuilder.Build(ctx, targetDir, artifacts, input.Options)
	if err != nil {
		return &CommandError{
			Code: 1,
			Diagnostics: []*diagnostics.Diagnostic{{
				Severity: diagnostics.SeverityError,
				Code:     diagnostics.CodeGenConflict,
				Message:  fmt.Sprintf("conflict building plan: %v", err),
				Hint:     "use --force to overwrite existing files",
				File:     targetDir,
			}},
		}
	}

	if opts.dryRun {
		if globalOpts.JSON {
			enc := json.NewEncoder(cmd.OutOrStdout())
			enc.SetIndent("", "  ")
			return enc.Encode(executionPlan)
		}
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Plan operations (dry-run) for %s %s:\n", gen.Name(), name)
		for _, op := range executionPlan.Operations {
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "  [%s] %s\n", op.Type, op.Path)
		}
		return nil
	}

	executor := plan.NewExecutor(fs)
	if err := executor.Execute(ctx, executionPlan); err != nil {
		return &CommandError{
			Code: 1,
			Diagnostics: []*diagnostics.Diagnostic{{
				Severity: diagnostics.SeverityError,
				Code:     diagnostics.CodeGenExecutionError,
				Message:  fmt.Sprintf("executing plan: %v", err),
				File:     targetDir,
			}},
		}
	}

	if globalOpts.JSON {
		enc := json.NewEncoder(cmd.OutOrStdout())
		enc.SetIndent("", "  ")
		return enc.Encode(executionPlan)
	}

	if !globalOpts.Quiet {
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Successfully generated %s %s\n", gen.Name(), name)
		for _, op := range executionPlan.Operations {
			if op.Type != plan.OpSkip {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "  + %s (%s)\n", op.Path, op.Type)
			}
		}
	}

	if runner != nil && !opts.dryRun {
		_, _ = runner.Run(ctx, targetDir, "go", "mod", "tidy")
	}

	return nil
}

func checkSQLC(cmd *cobra.Command, fs filesystem.FileSystem, runner process.Runner, opts *makeOptions) {
	ctx := cmd.Context()
	globalOpts := GetOptions(ctx)

	targetDir, _, err := resolveProjectTarget(ctx, fs, runner, opts.target)
	if err != nil {
		targetDir = "."
	}

	_, err = exec.LookPath("sqlc")
	if err != nil {
		if !globalOpts.JSON && !globalOpts.Quiet {
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), "\nNote: 'sqlc' binary was not found on $PATH.")
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), "To generate type-safe Go SQL queries, install it with:")
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), "  go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest")
		}
		return
	}

	if runner != nil {
		_, _ = runner.Run(ctx, targetDir, "sqlc", "generate")
	}
}

func newViewCmd(fs filesystem.FileSystem, runner process.Runner, opts *makeOptions) *cobra.Command {
	var partial bool
	cmd := &cobra.Command{
		Use:   "view <name>",
		Short: "Scaffold Templ view component or page",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			ctx := cmd.Context()
			globalOpts := GetOptions(ctx)

			targetDir, modulePath, err := resolveProjectTarget(ctx, fs, runner, opts.target)
			if err != nil {
				return err
			}

			gen := builtin.NewViewGenerator(modulePath)
			inputArgs := make(map[string]string)
			if partial {
				inputArgs["partial"] = "true"
			}

			input := generator.Input{
				Name: name,
				Args: inputArgs,
				Options: generator.Options{
					Force:  opts.force,
					DryRun: opts.dryRun,
				},
			}

			artifacts, err := gen.Generate(ctx, input)
			if err != nil {
				return &CommandError{
					Code: 1,
					Diagnostics: []*diagnostics.Diagnostic{{
						Severity: diagnostics.SeverityError,
						Code:     diagnostics.CodeGenExecutionError,
						Message:  fmt.Sprintf("generating view %s: %v", name, err),
						File:     targetDir,
					}},
				}
			}

			planBuilder := plan.NewBuilder(fs)
			executionPlan, err := planBuilder.Build(ctx, targetDir, artifacts, input.Options)
			if err != nil {
				return &CommandError{
					Code: 1,
					Diagnostics: []*diagnostics.Diagnostic{{
						Severity: diagnostics.SeverityError,
						Code:     diagnostics.CodeGenConflict,
						Message:  fmt.Sprintf("conflict building plan: %v", err),
						Hint:     "use --force to overwrite existing files",
						File:     targetDir,
					}},
				}
			}

			if opts.dryRun {
				if globalOpts.JSON {
					enc := json.NewEncoder(cmd.OutOrStdout())
					enc.SetIndent("", "  ")
					return enc.Encode(executionPlan)
				}
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Plan operations (dry-run) for %s %s:\n", gen.Name(), name)
				for _, op := range executionPlan.Operations {
					_, _ = fmt.Fprintf(cmd.OutOrStdout(), "  [%s] %s\n", op.Type, op.Path)
				}
				return nil
			}

			executor := plan.NewExecutor(fs)
			if err := executor.Execute(ctx, executionPlan); err != nil {
				return &CommandError{
					Code: 1,
					Diagnostics: []*diagnostics.Diagnostic{{
						Severity: diagnostics.SeverityError,
						Code:     diagnostics.CodeGenExecutionError,
						Message:  fmt.Sprintf("executing plan: %v", err),
						File:     targetDir,
					}},
				}
			}

			if !globalOpts.Quiet {
				if globalOpts.JSON {
					_, _ = fmt.Fprintf(cmd.OutOrStdout(), `{"status":"generated","artifact":%q,"name":%q,"path":%q}`+"\n", gen.Name(), name, artifacts[0].Path)
				} else {
					_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Generated %s: %s\n", gen.Name(), artifacts[0].Path)
				}
			}

			return nil
		},
	}
	cmd.Flags().BoolVar(&partial, "partial", false, "Scaffold as partial UI component instead of full page")
	return cmd
}

func newDockerCmd(fs filesystem.FileSystem, runner process.Runner, opts *makeOptions) *cobra.Command {
	var target string
	cmd := &cobra.Command{
		Use:   "docker [name]",
		Short: "Scaffold production multi-stage Dockerfile and docker-compose.yml",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := "docker"
			if len(args) > 0 {
				name = args[0]
			}
			return runDeployArtifact(cmd, fs, runner, opts, func(mod string, man *manifest.Manifest) generator.Generator {
				gen := builtin.NewDockerGenerator(mod)
				if man != nil {
					gen.WithCapabilities(man.Defaults.Database != "", man.Defaults.Cache != "", man.Defaults.Queue != "", man.Defaults.Assets == "vite")
					if target != "" {
						gen.WithTarget(target)
					} else if man.Defaults.Template != "" || man.Defaults.Assets == "vite" {
						gen.WithTarget("web")
					}
					if len(man.Workspace.Apps) > 0 || man.Workspace.DefaultTarget != "" {
						gen.WithWorkspace(true)
					}
				} else if target != "" {
					gen.WithTarget(target)
				}
				return gen
			}, name)
		},
	}
	cmd.Flags().StringVar(&target, "target-binary", "", "Target command binary to compile (e.g. api, web, worker)")
	return cmd
}

func newK8sCmd(fs filesystem.FileSystem, runner process.Runner, opts *makeOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "k8s [name]",
		Short: "Scaffold cloud-native Kubernetes manifests (deploy/k8s/)",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := "k8s"
			if len(args) > 0 {
				name = args[0]
			}
			return runDeployArtifact(cmd, fs, runner, opts, func(mod string, man *manifest.Manifest) generator.Generator {
				gen := builtin.NewK8sGenerator(mod)
				if man != nil {
					gen.WithCapabilities(man.Defaults.Database != "", man.Defaults.Cache != "", man.Defaults.Queue != "")
				}
				return gen
			}, name)
		},
	}
	return cmd
}

func newHelmCmd(fs filesystem.FileSystem, runner process.Runner, opts *makeOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "helm [name]",
		Short: "Scaffold Helm chart for application (deploy/helm/<name>/)",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := "helm"
			if len(args) > 0 {
				name = args[0]
			}
			return runDeployArtifact(cmd, fs, runner, opts, func(mod string, man *manifest.Manifest) generator.Generator {
				gen := builtin.NewHelmGenerator(mod)
				if man != nil {
					gen.WithCapabilities(man.Defaults.Database != "", man.Defaults.Cache != "", man.Defaults.Queue != "")
				}
				return gen
			}, name)
		},
	}
	return cmd
}

func newCICmd(fs filesystem.FileSystem, runner process.Runner, opts *makeOptions) *cobra.Command {
	var provider string
	cmd := &cobra.Command{
		Use:   "ci [name]",
		Short: "Scaffold CI/CD pipeline workflow (GitHub Actions or GitLab CI)",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := "ci"
			if len(args) > 0 {
				name = args[0]
			}
			return runDeployArtifact(cmd, fs, runner, opts, func(mod string, man *manifest.Manifest) generator.Generator {
				gen := builtin.NewCIGenerator(mod).WithProvider(provider)
				if man != nil {
					gen.WithCapabilities(man.Defaults.Database != "" && man.Defaults.Database != "none", man.Defaults.Cache != "" && man.Defaults.Cache != "none", man.Defaults.Queue != "" && man.Defaults.Queue != "none")
					gen.WithDatabaseDialect(man.Defaults.Database)
				}
				return gen
			}, name)
		},
	}
	cmd.Flags().StringVar(&provider, "provider", "github", "CI provider (github or gitlab)")
	return cmd
}

func newDeployCmd(fs filesystem.FileSystem, runner process.Runner, opts *makeOptions) *cobra.Command {
	var (
		targetBinary string
		ciProvider   string
	)
	cmd := &cobra.Command{
		Use:   "deploy [type]",
		Short: "Scaffold production deployment assets (docker, k8s, helm, ci, or all)",
		Long: `Scaffold deployment and infrastructure assets:
  - docker: Multi-stage Dockerfile & docker-compose.yml
  - k8s:    Kubernetes manifests in deploy/k8s/
  - helm:   Helm chart in deploy/helm/<app>/
  - ci:     CI pipeline workflow (.github/workflows/ci.yml)
  - all:    Scaffold all deployment targets together (default)`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			deployType := "all"
			if len(args) > 0 {
				deployType = strings.ToLower(args[0])
			}

			var generators []func(mod string, man *manifest.Manifest) generator.Generator
			var names []string

			switch deployType {
			case "docker":
				generators = append(generators, func(mod string, man *manifest.Manifest) generator.Generator {
					g := builtin.NewDockerGenerator(mod)
					if man != nil {
						g.WithCapabilities(man.Defaults.Database != "", man.Defaults.Cache != "", man.Defaults.Queue != "", man.Defaults.Assets == "vite")
						if targetBinary != "" {
							g.WithTarget(targetBinary)
						} else if man.Defaults.Template != "" || man.Defaults.Assets == "vite" {
							g.WithTarget("web")
						}
						if len(man.Workspace.Apps) > 0 || man.Workspace.DefaultTarget != "" {
							g.WithWorkspace(true)
						}
					}
					return g
				})
				names = append(names, "docker")
			case "k8s":
				generators = append(generators, func(mod string, man *manifest.Manifest) generator.Generator {
					g := builtin.NewK8sGenerator(mod)
					if man != nil {
						g.WithCapabilities(man.Defaults.Database != "", man.Defaults.Cache != "", man.Defaults.Queue != "")
					}
					return g
				})
				names = append(names, "k8s")
			case "helm":
				generators = append(generators, func(mod string, man *manifest.Manifest) generator.Generator {
					g := builtin.NewHelmGenerator(mod)
					if man != nil {
						g.WithCapabilities(man.Defaults.Database != "", man.Defaults.Cache != "", man.Defaults.Queue != "")
					}
					return g
				})
				names = append(names, "helm")
			case "ci":
				generators = append(generators, func(mod string, man *manifest.Manifest) generator.Generator {
					g := builtin.NewCIGenerator(mod).WithProvider(ciProvider)
					if man != nil {
						g.WithCapabilities(man.Defaults.Database != "" && man.Defaults.Database != "none", man.Defaults.Cache != "" && man.Defaults.Cache != "none", man.Defaults.Queue != "" && man.Defaults.Queue != "none")
						g.WithDatabaseDialect(man.Defaults.Database)
					}
					return g
				})
				names = append(names, "ci")
			case "all":
				generators = []func(mod string, man *manifest.Manifest) generator.Generator{
					func(mod string, man *manifest.Manifest) generator.Generator {
						g := builtin.NewDockerGenerator(mod)
						if man != nil {
							g.WithCapabilities(man.Defaults.Database != "", man.Defaults.Cache != "", man.Defaults.Queue != "", man.Defaults.Assets == "vite")
							if targetBinary != "" {
								g.WithTarget(targetBinary)
							} else if man.Defaults.Template != "" || man.Defaults.Assets == "vite" {
								g.WithTarget("web")
							}
								if len(man.Workspace.Apps) > 0 || man.Workspace.DefaultTarget != "" {
									g.WithWorkspace(true)
								}
						}
						return g
					},
					func(mod string, man *manifest.Manifest) generator.Generator {
						g := builtin.NewK8sGenerator(mod)
						if man != nil {
							g.WithCapabilities(man.Defaults.Database != "", man.Defaults.Cache != "", man.Defaults.Queue != "")
						}
						return g
					},
					func(mod string, man *manifest.Manifest) generator.Generator {
						g := builtin.NewHelmGenerator(mod)
						if man != nil {
							g.WithCapabilities(man.Defaults.Database != "", man.Defaults.Cache != "", man.Defaults.Queue != "")
						}
						return g
					},
					func(mod string, man *manifest.Manifest) generator.Generator {
						g := builtin.NewCIGenerator(mod).WithProvider(ciProvider)
						if man != nil {
							g.WithCapabilities(man.Defaults.Database != "" && man.Defaults.Database != "none", man.Defaults.Cache != "" && man.Defaults.Cache != "none", man.Defaults.Queue != "" && man.Defaults.Queue != "none")
							g.WithDatabaseDialect(man.Defaults.Database)
						}
						return g
					},
				}
				names = []string{"docker", "k8s", "helm", "ci"}
			default:
				return fmt.Errorf("unknown deploy target %q: valid targets are all, docker, k8s, helm, ci", deployType)
			}

			targetDir, modulePath, err := resolveProjectTarget(cmd.Context(), fs, runner, opts.target)
			if err != nil {
				return err
			}

			var man *manifest.Manifest
			mPath := filepath.Join(targetDir, "loy.yaml")
			if mData, err := fs.ReadFile(mPath); err == nil {
				parser := manifest.NewParser()
				if m, diag := parser.ParseStrict(mPath, mData); diag == nil {
					man = m
				}
			}

			var allArtifacts []model.Artifact
			for i, gFn := range generators {
				gen := gFn(modulePath, man)
				input := generator.Input{
					Name: names[i],
					Options: generator.Options{
						Force:  opts.force,
						DryRun: opts.dryRun,
					},
				}
				arts, err := gen.Generate(cmd.Context(), input)
				if err != nil {
					return &CommandError{
						Code: 1,
						Diagnostics: []*diagnostics.Diagnostic{{
							Severity: diagnostics.SeverityError,
							Code:     diagnostics.CodeGenExecutionError,
							Message:  fmt.Sprintf("generating %s: %v", gen.Name(), err),
							File:     targetDir,
						}},
					}
				}
				allArtifacts = append(allArtifacts, arts...)
			}

			return executeDeployPlan(cmd, fs, opts, targetDir, allArtifacts, deployType)
		},
	}
	cmd.Flags().StringVar(&targetBinary, "target-binary", "", "Target command binary to compile for Dockerfile")
	cmd.Flags().StringVar(&ciProvider, "provider", "github", "CI provider (github or gitlab)")
	return cmd
}

func runDeployArtifact(cmd *cobra.Command, fs filesystem.FileSystem, runner process.Runner, opts *makeOptions, factory func(mod string, man *manifest.Manifest) generator.Generator, name string) error {
	ctx := cmd.Context()

	targetDir, modulePath, err := resolveProjectTarget(ctx, fs, runner, opts.target)
	if err != nil {
		return err
	}

	var man *manifest.Manifest
	mPath := filepath.Join(targetDir, "loy.yaml")
	if mData, err := fs.ReadFile(mPath); err == nil {
		parser := manifest.NewParser()
		if m, diag := parser.ParseStrict(mPath, mData); diag == nil {
			man = m
		}
	}

	gen := factory(modulePath, man)
	input := generator.Input{
		Name: name,
		Options: generator.Options{
			Force:  opts.force,
			DryRun: opts.dryRun,
		},
	}

	artifacts, err := gen.Generate(ctx, input)
	if err != nil {
		return &CommandError{
			Code: 1,
			Diagnostics: []*diagnostics.Diagnostic{{
				Severity: diagnostics.SeverityError,
				Code:     diagnostics.CodeGenExecutionError,
				Message:  fmt.Sprintf("generating %s: %v", gen.Name(), err),
				File:     targetDir,
			}},
		}
	}

	return executeDeployPlan(cmd, fs, opts, targetDir, artifacts, gen.Name())
}

func executeDeployPlan(cmd *cobra.Command, fs filesystem.FileSystem, opts *makeOptions, targetDir string, artifacts []model.Artifact, title string) error {
	ctx := cmd.Context()
	globalOpts := GetOptions(ctx)

	planBuilder := plan.NewBuilder(fs)
	executionPlan, err := planBuilder.Build(ctx, targetDir, artifacts, generator.Options{
		Force:  opts.force,
		DryRun: opts.dryRun,
	})
	if err != nil {
		return &CommandError{
			Code: 1,
			Diagnostics: []*diagnostics.Diagnostic{{
				Severity: diagnostics.SeverityError,
				Code:     diagnostics.CodeGenConflict,
				Message:  fmt.Sprintf("conflict building plan: %v", err),
				Hint:     "use --force to overwrite existing files",
				File:     targetDir,
			}},
		}
	}

	if opts.dryRun {
		if globalOpts.JSON {
			enc := json.NewEncoder(cmd.OutOrStdout())
			enc.SetIndent("", "  ")
			return enc.Encode(executionPlan)
		}
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Plan operations (dry-run) for %s:\n", title)
		for _, op := range executionPlan.Operations {
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "  [%s] %s\n", op.Type, op.Path)
		}
		return nil
	}

	executor := plan.NewExecutor(fs)
	if err := executor.Execute(ctx, executionPlan); err != nil {
		return &CommandError{
			Code: 1,
			Diagnostics: []*diagnostics.Diagnostic{{
				Severity: diagnostics.SeverityError,
				Code:     diagnostics.CodeGenExecutionError,
				Message:  fmt.Sprintf("executing plan: %v", err),
				File:     targetDir,
			}},
		}
	}

	if !globalOpts.Quiet {
		if globalOpts.JSON {
			paths := make([]string, len(artifacts))
			for i, a := range artifacts {
				paths[i] = a.Path
			}
			data, _ := json.Marshal(map[string]any{
				"status":    "generated",
				"deploy":    title,
				"artifacts": paths,
			})
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), string(data))
		} else {
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Generated %s (%d files):\n", title, len(artifacts))
			for _, a := range artifacts {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "  + %s\n", a.Path)
			}
		}
	}

	return nil
}
