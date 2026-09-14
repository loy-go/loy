package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/uloydev/loy/internal/architecture"
	"github.com/uloydev/loy/internal/architecture/rules"
	"github.com/uloydev/loy/internal/diagnostics"
	"github.com/uloydev/loy/internal/discovery"
	"github.com/uloydev/loy/internal/filesystem"
	"github.com/uloydev/loy/internal/manifest"
	"github.com/uloydev/loy/internal/process"
	"golang.org/x/mod/modfile"
)

func newCheckCmd(fs filesystem.FileSystem, runner process.Runner) *cobra.Command {
	var (
		deep   bool
		strict bool
	)

	cmd := &cobra.Command{
		Use:   "check [path]",
		Short: "Validate architectural rules and boundaries",
		Long:  `Validates that Go code strictly adheres to Loy architectural invariants, layer boundaries, and workspace topology.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			opts := GetOptions(ctx)

			targetDir := "."
			if len(args) > 0 {
				targetDir = args[0]
			}

			// Discover project root and manifest
			disc, err := discovery.NewDiscoverer(fs, runner)
			if err != nil {
				return &CommandError{
					Code: 1,
					Diagnostics: []*diagnostics.Diagnostic{
						{
							Severity: diagnostics.SeverityError,
							Code:     diagnostics.CodeCLIExecutionFail,
							Message:  fmt.Sprintf("initializing discoverer: %v", err),
						},
					},
				}
			}

			discRes, diag := disc.Discover(ctx, targetDir)
			if diag != nil {
				return &CommandError{
					Code: 1,
					Diagnostics: []*diagnostics.Diagnostic{diag},
				}
			}

			// Load loy.yaml if present
			var mf manifest.Manifest
			manifestPath := discRes.ManifestPath
			if discRes.HasManifest {
				data, readErr := fs.ReadFile(manifestPath)
				if readErr == nil {
					p := manifest.NewParser()
					parsed, _ := p.ParseStrict(manifestPath, data)
					if parsed != nil {
						mf = *parsed
					}
				}
			}

			moduleName := mf.Project.Module
			if moduleName == "" && discRes.HasGoMod {
				data, readErr := fs.ReadFile(discRes.GoModPath)
				if readErr == nil {
					f, parseErr := modfile.Parse(discRes.GoModPath, data, nil)
					if parseErr == nil && f.Module != nil {
						moduleName = f.Module.Mod.Path
					}
				}
			}

			cfg := architecture.AnalyzerConfig{
				ModuleName: moduleName,
				RootDir:    discRes.RootDir,
				Deep:       deep,
				Strict:     strict || mf.Architecture.Strict,
				Rules:      rules.DefaultRules(),
			}

			analyzer := architecture.NewAnalyzer(fs, cfg)
			violations, err := analyzer.Run(ctx)
			if err != nil {
				return &CommandError{
					Code: 3,
					Diagnostics: []*diagnostics.Diagnostic{
						{
							Severity: diagnostics.SeverityError,
							Code:     diagnostics.CodeInternalError,
							Message:  fmt.Sprintf("analysis failed: %v", err),
						},
					},
				}
			}

			var cmdDiags []*diagnostics.Diagnostic
			for _, v := range violations {
				d := v.ToDiagnostic()
				cmdDiags = append(cmdDiags, &d)
			}

			if len(violations) > 0 {
				return &CommandError{
					Code:        1,
					Diagnostics: cmdDiags,
				}
			}

			if opts.JSON && len(violations) == 0 {
				_ = (&diagnostics.JSONFormatter{Indent: true}).Format(cmd.OutOrStdout(), []diagnostics.Diagnostic{})
			} else if !opts.Quiet && !opts.JSON {
				fmt.Fprintln(cmd.OutOrStdout(), "All architecture rules passed.")
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&deep, "deep", false, "Run deep type analysis via go/packages")
	cmd.Flags().BoolVar(&strict, "strict", false, "Treat all architectural warnings as fatal errors")

	return cmd
}
