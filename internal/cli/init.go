package cli

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/loy-go/loy/internal/diagnostics"
	"github.com/loy-go/loy/internal/discovery"
	"github.com/loy-go/loy/internal/filesystem"
	"github.com/loy-go/loy/internal/preset"
	"github.com/loy-go/loy/internal/process"
)

func newInitCmd(fs filesystem.FileSystem, runner process.Runner) *cobra.Command {
	var presetName string
	var projectName string
	var force bool

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize Loy configuration in an existing Go project",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			opts := GetOptions(ctx)

			currDir := "."

			disc, err := discovery.NewDiscoverer(fs, runner)
			if err != nil {
				return err
			}

			discovered, _ := disc.Discover(ctx, currDir)
			targetDir := currDir
			if discovered != nil && discovered.RootDir != "" {
				targetDir = discovered.RootDir
			}

			cleanTarget, pathErr := filesystem.CleanAndValidatePath(targetDir, targetDir)
			if pathErr != nil {
				return &CommandError{
					Code: 1,
					Diagnostics: []*diagnostics.Diagnostic{{
						Severity: diagnostics.SeverityError,
						Code:     diagnostics.CodeFSPathTraversal,
						Message:  fmt.Sprintf("validating target directory: %v", pathErr),
						File:     targetDir,
					}},
				}
			}

			// Must detect go.mod in target directory per Phase 2 Spec
			goModPath, _ := filesystem.CleanAndValidatePath(cleanTarget, filepath.Join(cleanTarget, "go.mod"))
			hasMod, _ := fs.Exists(goModPath)
			if !hasMod {
				return &CommandError{
					Code: 1,
					Diagnostics: []*diagnostics.Diagnostic{{
						Severity: diagnostics.SeverityError,
						Code:     diagnostics.CodeProjectRootNotFound,
						Message:  "no go.mod found in target directory; 'loy init' requires an existing Go module",
						Hint:     "run 'go mod init <module>' first or use 'loy new <name>' to create a fresh project",
						File:     cleanTarget,
					}},
				}
			}

			manifestPath, _ := filesystem.CleanAndValidatePath(cleanTarget, filepath.Join(cleanTarget, "loy.yaml"))
			exists, _ := fs.Exists(manifestPath)
			if exists && !force {
				return &CommandError{
					Code: 1,
					Diagnostics: []*diagnostics.Diagnostic{{
						Severity: diagnostics.SeverityError,
						Code:     diagnostics.CodeProjectAlreadyExists,
						Message:  "loy.yaml already exists in target directory",
						Hint:     "use --force to overwrite existing loy.yaml",
						File:     manifestPath,
					}},
				}
			}

			// Infer project name if not specified
			if projectName == "" {
				if inferred, err := disc.InferProjectName(cleanTarget); err == nil && inferred != "" {
					projectName = inferred
				} else {
					projectName = filepath.Base(cleanTarget)
					if projectName == "." || projectName == "/" || projectName == "" {
						projectName = "myapp"
					}
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
						File:     manifestPath,
					}},
				}
			}

			yamlContent := p.MaterializeYAML(projectName)
			if err := fs.WriteFile(manifestPath, []byte(yamlContent), 0644); err != nil {
				return fmt.Errorf("writing loy.yaml: %w", err)
			}

			if !opts.Quiet {
				if opts.JSON {
					_, _ = fmt.Fprintf(cmd.OutOrStdout(), `{"status":"initialized","file":%q,"preset":%q,"project":%q}`+"\n", manifestPath, presetName, projectName)
				} else {
					_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Initialized Loy project %q using preset %q in %s\n", projectName, presetName, manifestPath)
				}
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&presetName, "preset", "p", "api", "Preset to initialize (api, fullstack, minimal, monorepo, web)")
	cmd.Flags().StringVarP(&projectName, "name", "n", "", "Project name override")
	cmd.Flags().BoolVarP(&force, "force", "f", false, "Overwrite existing loy.yaml")

	return cmd
}
