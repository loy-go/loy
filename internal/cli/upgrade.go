package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/loy-go/loy/internal/diagnostics"
	"github.com/loy-go/loy/internal/discovery"
	"github.com/loy-go/loy/internal/filesystem"
	"github.com/loy-go/loy/internal/process"
	"gopkg.in/yaml.v3"
)

// UpgradeAction represents a planned or applied modification to configuration or managed files.
type UpgradeAction struct {
	Path        string `json:"path"`
	Action      string `json:"action"` // "update_config", "restore_region", etc.
	Description string `json:"description"`
	Content     []byte `json:"-"`
}

// UpgradeResult represents the machine-readable summary of loy upgrade.
type UpgradeResult struct {
	Status  string          `json:"status"` // "up_to_date", "dry_run", "applied"
	Actions []UpgradeAction `json:"actions"`
}

func newUpgradeCmd(fs filesystem.FileSystem, runner process.Runner) *cobra.Command {
	var (
		dryRun bool
		force  bool
	)

	cmd := &cobra.Command{
		Use:   "upgrade [path]",
		Short: "Migrate project configuration and restore managed regions without touching domain logic",
		Long: `loy upgrade inspects project configuration (loy.yaml) and managed artifacts (e.g. internal/app/wiring.go),
planning and non-destructively applying schema upgrades and missing managed region markers.
Developer-owned domain and service logic is never rewritten.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			targetDir := "."
			if len(args) > 0 {
				targetDir = args[0]
			}
			ctx := cmd.Context()
			opts := GetOptions(ctx)

			disc, err := discovery.NewDiscoverer(fs, runner)
			if err != nil {
				return &CommandError{
					Code: 1,
					Diagnostics: []*diagnostics.Diagnostic{{
						Severity: diagnostics.SeverityError,
						Code:     diagnostics.CodeInternalError,
						Message:  fmt.Sprintf("initializing discoverer: %v", err),
					}},
				}
			}

			discRes, diag := disc.Discover(ctx, targetDir)
			if diag != nil {
				return &CommandError{Code: 1, Diagnostics: []*diagnostics.Diagnostic{diag}}
			}
			if discRes == nil || !discRes.HasManifest {
				return &CommandError{
					Code: 1,
					Diagnostics: []*diagnostics.Diagnostic{{
						Severity: diagnostics.SeverityError,
						Code:     diagnostics.CodeProjectRootNotFound,
						Message:  "loy.yaml not found in project hierarchy",
						Hint:     "run 'loy init' or 'loy new' to initialize a Loy project",
					}},
				}
			}

			projectRoot := discRes.RootDir
			actions, err := planProjectUpgrades(ctx, fs, projectRoot, discRes.ManifestPath)
			if err != nil {
				return &CommandError{
					Code: 1,
					Diagnostics: []*diagnostics.Diagnostic{{
						Severity: diagnostics.SeverityError,
						Code:     diagnostics.CodeConfigValidationError,
						Message:  fmt.Sprintf("planning upgrades: %v", err),
					}},
				}
			}

			if len(actions) == 0 {
				if opts.JSON {
					return json.NewEncoder(cmd.OutOrStdout()).Encode(UpgradeResult{
						Status:  "up_to_date",
						Actions: []UpgradeAction{},
					})
				}
				if !opts.Quiet {
					_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Project configuration and managed artifacts are up to date (Loy schema v1).")
				}
				return nil
			}

			if dryRun {
				if opts.JSON {
					return json.NewEncoder(cmd.OutOrStdout()).Encode(UpgradeResult{
						Status:  "dry_run",
						Actions: actions,
					})
				}
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Proposed upgrade plan for %s:\n\n", projectRoot)
				for _, a := range actions {
					_, _ = fmt.Fprintf(cmd.OutOrStdout(), "  [%s] %s: %s\n", a.Action, a.Path, a.Description)
				}
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "\n[Dry Run] %d upgrade action(s) planned. Run 'loy upgrade' to apply.\n", len(actions))
				return nil
			}

			// Apply actions safely
			for _, a := range actions {
				absPath, err := filesystem.CleanAndValidatePath(projectRoot, a.Path)
				if err != nil {
					return fmt.Errorf("validating path %s: %w", a.Path, err)
				}
				if a.Content == nil {
					continue
				}
				if err := fs.WriteFile(absPath, a.Content, 0644); err != nil {
					return fmt.Errorf("applying upgrade to %s: %w", a.Path, err)
				}
			}

			if opts.JSON {
				return json.NewEncoder(cmd.OutOrStdout()).Encode(UpgradeResult{
					Status:  "applied",
					Actions: actions,
				})
			}

			if !opts.Quiet {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Successfully applied %d upgrade(s) to %s:\n", len(actions), projectRoot)
				for _, a := range actions {
					_, _ = fmt.Fprintf(cmd.OutOrStdout(), "  ✓ %s (%s)\n", a.Path, a.Description)
				}
			}

			return nil
		},
	}

	cmd.Flags().BoolVarP(&dryRun, "dry-run", "n", false, "Preview proposed upgrades without modifying files")
	cmd.Flags().BoolVarP(&force, "force", "f", false, "Force apply upgrades even if non-critical warnings exist")

	return cmd
}

func planProjectUpgrades(ctx context.Context, fs filesystem.FileSystem, projectRoot, manifestPath string) ([]UpgradeAction, error) {
	var actions []UpgradeAction

	// 1. Inspect and upgrade loy.yaml
	mData, err := fs.ReadFile(manifestPath)
	if err != nil {
		return nil, fmt.Errorf("reading loy.yaml: %w", err)
	}

	var rawMap map[string]any
	if err := yaml.Unmarshal(mData, &rawMap); err != nil {
		return nil, fmt.Errorf("parsing loy.yaml: %w", err)
	}

	if rawMap != nil {
		modified := false
		relPath, _ := filepath.Rel(projectRoot, manifestPath)
		if relPath == "" {
			relPath = "loy.yaml"
		}

		v, hasVersion := rawMap["version"]
		if !hasVersion || v == 0 || v == nil {
			rawMap["version"] = 1
			modified = true
			actions = append(actions, UpgradeAction{
				Path:        relPath,
				Action:      "update_config",
				Description: "Set configuration schema to version 1",
			})
		}

		if _, hasDefaults := rawMap["defaults"]; !hasDefaults {
			// Provide standard defaults
			rawMap["defaults"] = map[string]string{
				"http":     "fiber",
				"database": "postgres",
				"cache":    "valkey",
				"queue":    "asynq",
			}
			modified = true
			actions = append(actions, UpgradeAction{
				Path:        relPath,
				Action:      "update_config",
				Description: "Populate missing defaults section in loy.yaml",
			})
		}

		if modified {
			var outBuf bytes.Buffer
			enc := yaml.NewEncoder(&outBuf)
			enc.SetIndent(2)
			_ = enc.Encode(rawMap)
			// Attach content to all actions for this file
			for i := range actions {
				if actions[i].Path == relPath {
					actions[i].Content = outBuf.Bytes()
				}
			}
		}
	}

	// 2. Inspect managed regions in internal/app/wiring.go
	wiringPath, _ := filesystem.CleanAndValidatePath(projectRoot, filepath.Join(projectRoot, "internal/app/wiring.go"))
	if exists, _ := fs.Exists(wiringPath); exists {
		wBytes, err := fs.ReadFile(wiringPath)
		if err == nil {
			wContent := string(wBytes)
			requiredRegions := []string{"imports", "repositories", "services", "handlers", "routes"}
			var missingRegions []string
			for _, r := range requiredRegions {
				marker := fmt.Sprintf("// loy:region:%s", r)
				if !strings.Contains(wContent, marker) {
					missingRegions = append(missingRegions, r)
				}
			}

			if len(missingRegions) > 0 {
				newWiringContent := restoreWiringRegions(wContent, missingRegions)
				actions = append(actions, UpgradeAction{
					Path:        "internal/app/wiring.go",
					Action:      "restore_region",
					Description: fmt.Sprintf("Restore missing managed comment region(s): %s", strings.Join(missingRegions, ", ")),
					Content:     []byte(newWiringContent),
				})
			}
		}
	}

	return actions, nil
}

func restoreWiringRegions(content string, missing []string) string {
	for _, m := range missing {
		switch m {
		case "imports":
			if !strings.Contains(content, "// loy:region:imports") {
				content = strings.Replace(content, "import (", "import (\n\t// loy:region:imports\n\t// loy:endregion", 1)
			}
		case "repositories":
			if !strings.Contains(content, "// loy:region:repositories") {
				marker := "\n\t// loy:region:repositories\n\t// loy:endregion\n"
				content = strings.Replace(content, "func (a *App) wireDependencies() error {", "func (a *App) wireDependencies() error {"+marker, 1)
			}
		case "services":
			if !strings.Contains(content, "// loy:region:services") {
				marker := "\t// loy:region:services\n\t// loy:endregion\n\n"
				if strings.Contains(content, "// loy:region:handlers") {
					content = strings.Replace(content, "// loy:region:handlers", marker+"\t// loy:region:handlers", 1)
				} else if strings.Contains(content, "// loy:region:routes") {
					content = strings.Replace(content, "// loy:region:routes", marker+"\t// loy:region:routes", 1)
				} else if strings.Contains(content, "return nil") {
					content = strings.Replace(content, "return nil", marker+"\treturn nil", 1)
				}
			}
		case "handlers":
			if !strings.Contains(content, "// loy:region:handlers") {
				marker := "\t// loy:region:handlers\n\t// loy:endregion\n\n"
				if strings.Contains(content, "// loy:region:routes") {
					content = strings.Replace(content, "// loy:region:routes", marker+"\t// loy:region:routes", 1)
				} else if strings.Contains(content, "return nil") {
					content = strings.Replace(content, "return nil", marker+"\treturn nil", 1)
				}
			}
		case "routes":
			if !strings.Contains(content, "// loy:region:routes") {
				marker := "\t// loy:region:routes\n\t// loy:endregion\n\n"
				content = strings.Replace(content, "return nil", marker+"\treturn nil", 1)
			}
		}
	}
	return content
}
