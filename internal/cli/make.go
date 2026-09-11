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
	"github.com/uloydev/loy/internal/diagnostics"
	"github.com/uloydev/loy/internal/discovery"
	"github.com/uloydev/loy/internal/filesystem"
	"github.com/uloydev/loy/internal/generator"
	"github.com/uloydev/loy/internal/generator/builtin"
	"github.com/uloydev/loy/internal/generator/builtin/wiring"
	"github.com/uloydev/loy/internal/generator/plan"
	"github.com/uloydev/loy/internal/process"
	"github.com/uloydev/loy/internal/workspace"
)

type makeOptions struct {
	target string
	force  bool
	dryRun bool
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

	cmd.AddCommand(newArtifactCmd("test", "Scaffold unit and integration tests", fs, runner, opts, func(mod string) generator.Generator {
		return builtin.NewTestGenerator(mod)
	}, []string{}))

	// Composers
	cmd.AddCommand(newFeatureCmd(fs, runner, opts))
	cmd.AddCommand(newCRUDCmd(fs, runner, opts))

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

	// Infer module name
	if inferred, infErr := disc.InferProjectName(targetDir); infErr == nil && inferred != "" {
		modulePath = inferred
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

	// If requires wiring, ensure wiring.go exists
	if requiresWiring && !opts.dryRun {
		splicerMgr := wiring.NewSplicerManager(fs)
		if err := splicerMgr.EnsureWiringFile(ctx, targetDir); err != nil {
			return fmt.Errorf("ensuring wiring.go: %w", err)
		}
	}

	gen := factory(modulePath)
	input := generator.Input{
		Name: name,
		Args: map[string]string{
			"fields": fields,
		},
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
				Message:  fmt.Sprintf("generating %s %s: %v", gen.Name(), name, err),
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
		fmt.Fprintf(cmd.OutOrStdout(), "Plan operations (dry-run) for %s %s:\n", gen.Name(), name)
		for _, op := range executionPlan.Operations {
			fmt.Fprintf(cmd.OutOrStdout(), "  [%s] %s\n", op.Type, op.Path)
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
		fmt.Fprintf(cmd.OutOrStdout(), "Successfully generated %s %s\n", gen.Name(), name)
		for _, op := range executionPlan.Operations {
			if op.Type != plan.OpSkip {
				fmt.Fprintf(cmd.OutOrStdout(), "  + %s (%s)\n", op.Path, op.Type)
			}
		}
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
			fmt.Fprintln(cmd.OutOrStdout(), "\nNote: 'sqlc' binary was not found on $PATH.")
			fmt.Fprintln(cmd.OutOrStdout(), "To generate type-safe Go SQL queries, install it with:")
			fmt.Fprintln(cmd.OutOrStdout(), "  go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest")
		}
		return
	}

	if runner != nil {
		_, _ = runner.Run(ctx, targetDir, "sqlc", "generate")
	}
}
