package cli

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/loy-go/loy/internal/diagnostics"
	"github.com/loy-go/loy/internal/discovery"
	"github.com/loy-go/loy/internal/filesystem"
	"github.com/loy-go/loy/internal/manifest"
	"github.com/loy-go/loy/internal/plugin"
	"github.com/loy-go/loy/internal/process"
	"golang.org/x/mod/modfile"
)

func newPluginCmd(fs filesystem.FileSystem, runner process.Runner) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "plugin",
		Short: "Manage and execute sandboxed external community plugins and generators",
		Long: `The plugin command tree allows installing, inspecting, and running sandboxed community generators.
Plugins execute out-of-process and cannot write to disk directly; all emitted artifacts pass path jail validation and atomic plan execution.`,
	}

	cmd.AddCommand(newPluginListCmd(fs, runner))
	cmd.AddCommand(newPluginInstallCmd(fs, runner))
	cmd.AddCommand(newPluginRunCmd(fs, runner))

	return cmd
}

func newPluginListCmd(fs filesystem.FileSystem, runner process.Runner) *cobra.Command {
	return &cobra.Command{
		Use:   "list [path]",
		Short: "List installed community plugins in the project",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			targetDir := "."
			if len(args) > 0 {
				targetDir = args[0]
			} else if opts := GetOptions(cmd.Context()); opts.Directory != "" {
				targetDir = opts.Directory
			}
			ctx := cmd.Context()
			opts := GetOptions(ctx)

			disc, err := discovery.NewDiscoverer(fs, runner)
			if err == nil {
				if discRes, diag := disc.Discover(ctx, targetDir); diag == nil && discRes != nil {
					targetDir = discRes.RootDir
				}
			}

			mgr := plugin.NewManager(fs, runner, targetDir)
			plugins, err := mgr.List(ctx)
			if err != nil {
				return &CommandError{
					Code: 1,
					Diagnostics: []*diagnostics.Diagnostic{{
						Severity: diagnostics.SeverityError,
						Code:     diagnostics.CodeInternalError,
						Message:  fmt.Sprintf("listing plugins: %v", err),
					}},
				}
			}

			if opts.JSON {
				return json.NewEncoder(cmd.OutOrStdout()).Encode(plugins)
			}

			if len(plugins) == 0 {
				if !opts.Quiet {
					_, _ = fmt.Fprintln(cmd.OutOrStdout(), "No plugins installed in this project. Use 'loy plugin install <source>' to install one.")
				}
				return nil
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Installed Community Plugins (%d):\n\n", len(plugins))
			for _, p := range plugins {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "• %s (v%s) - %s\n", p.Name, p.Version, p.Description)
				for _, c := range p.Commands {
					_, _ = fmt.Fprintf(cmd.OutOrStdout(), "    - %s: %s\n", c.Name, c.Description)
				}
			}
			return nil
		},
	}
}

func newPluginInstallCmd(fs filesystem.FileSystem, runner process.Runner) *cobra.Command {
	return &cobra.Command{
		Use:   "install <source> [path]",
		Short: "Install a community plugin from a local directory or Git repository",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			source := args[0]
			ctx := cmd.Context()
			opts := GetOptions(ctx)
			targetDir := "."
			if len(args) > 1 {
				targetDir = args[1]
			} else if opts.Directory != "" {
				targetDir = opts.Directory
			}

			disc, err := discovery.NewDiscoverer(fs, runner)
			if err == nil {
				if discRes, diag := disc.Discover(ctx, targetDir); diag == nil && discRes != nil {
					targetDir = discRes.RootDir
				}
			}

			mgr := plugin.NewManager(fs, runner, targetDir)
			pm, err := mgr.Install(ctx, source)
			if err != nil {
				return &CommandError{
					Code: 1,
					Diagnostics: []*diagnostics.Diagnostic{{
						Severity: diagnostics.SeverityError,
						Code:     diagnostics.CodeCLIExecutionFail,
						Message:  fmt.Sprintf("installing plugin: %v", err),
					}},
				}
			}

			if opts.JSON {
				return json.NewEncoder(cmd.OutOrStdout()).Encode(map[string]any{
					"status":  "installed",
					"plugin":  pm.Name,
					"version": pm.Version,
				})
			}

			if !opts.Quiet {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Successfully installed plugin %q (v%s)\n", pm.Name, pm.Version)
				if len(pm.Commands) > 0 {
					_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Available commands:")
					for _, c := range pm.Commands {
						_, _ = fmt.Fprintf(cmd.OutOrStdout(), "  loy plugin run %s %s\n", pm.Name, c.Name)
					}
				}
			}
			return nil
		},
	}
}

func newPluginRunCmd(fs filesystem.FileSystem, runner process.Runner) *cobra.Command {
	var (
		dryRun bool
		force  bool
	)

	cmd := &cobra.Command{
		Use:                "run <plugin-name> <command> [args...]",
		Short:              "Execute an installed plugin generator command within the sandbox",
		Args:               cobra.MinimumNArgs(2),
		FParseErrWhitelist: cobra.FParseErrWhitelist{UnknownFlags: true},
		RunE: func(cmd *cobra.Command, args []string) error {
			pluginName := args[0]
			cmdName := args[1]
			extraArgs := args[2:]

			ctx := cmd.Context()
			opts := GetOptions(ctx)

			// Discover project root and module path
			targetDir := "."
			if opts.Directory != "" {
				targetDir = opts.Directory
			}
			moduleName := ""
			defaultsMap := make(map[string]string)

			disc, err := discovery.NewDiscoverer(fs, runner)
			if err == nil {
				if discRes, diag := disc.Discover(ctx, targetDir); diag == nil && discRes != nil {
					targetDir = discRes.RootDir
					if discRes.HasManifest {
						if data, err := fs.ReadFile(discRes.ManifestPath); err == nil {
							p := manifest.NewParser()
							if m, d := p.ParseStrict(discRes.ManifestPath, data); d == nil && m != nil {
								moduleName = m.Project.Module
								defaultsMap["http"] = m.Defaults.HTTP
								defaultsMap["database"] = m.Defaults.Database
								defaultsMap["cache"] = m.Defaults.Cache
								defaultsMap["queue"] = m.Defaults.Queue
							}
						}
					}
					if moduleName == "" && discRes.HasGoMod {
						if data, err := fs.ReadFile(discRes.GoModPath); err == nil {
							if f, err := modfile.Parse(discRes.GoModPath, data, nil); err == nil && f.Module != nil {
								moduleName = f.Module.Mod.Path
							}
						}
					}
				}
			}

			mgr := plugin.NewManager(fs, runner, targetDir)
			artifacts, err := mgr.Execute(ctx, pluginName, cmdName, extraArgs, moduleName, defaultsMap, dryRun, force)
			if err != nil {
				return &CommandError{
					Code: 1,
					Diagnostics: []*diagnostics.Diagnostic{{
						Severity: diagnostics.SeverityError,
						Code:     diagnostics.CodeGenExecutionError,
						Message:  fmt.Sprintf("executing plugin %s %s: %v", pluginName, cmdName, err),
					}},
				}
			}

			if opts.JSON {
				var paths []string
				for _, a := range artifacts {
					paths = append(paths, a.Path)
				}
				return json.NewEncoder(cmd.OutOrStdout()).Encode(map[string]any{
					"status":    "ok",
					"plugin":    pluginName,
					"command":   cmdName,
					"artifacts": paths,
				})
			}

			if !opts.Quiet {
				mode := "generated"
				if dryRun {
					mode = "planned (dry-run)"
				}
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Successfully %s %d artifact(s) via plugin %s %s:\n", mode, len(artifacts), pluginName, cmdName)
				for _, a := range artifacts {
					_, _ = fmt.Fprintf(cmd.OutOrStdout(), "  + %s\n", a.Path)
				}
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Preview generated artifacts without writing to disk")
	cmd.Flags().BoolVar(&force, "force", false, "Overwrite existing developer-owned files")

	return cmd
}
