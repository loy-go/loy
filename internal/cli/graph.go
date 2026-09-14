package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/loy-go/loy/internal/architecture"
	"github.com/loy-go/loy/internal/architecture/rules"
	"github.com/loy-go/loy/internal/discovery"
	"github.com/loy-go/loy/internal/filesystem"
	"github.com/loy-go/loy/internal/graph"
	"github.com/loy-go/loy/internal/manifest"
	"github.com/loy-go/loy/internal/process"
	"golang.org/x/mod/modfile"
)

func newGraphCmd(fs filesystem.FileSystem, runner process.Runner) *cobra.Command {
	var (
		format         string
		violationsOnly bool
	)

	cmd := &cobra.Command{
		Use:   "graph [path]",
		Short: "Visualize package dependency hierarchy and architectural layers",
		Long:  `loy graph renders project package dependencies, architectural layers, and boundary violations in ascii, dot, or mermaid format.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			targetDir := "."
			if len(args) > 0 {
				targetDir = args[0]
			}

			moduleName := ""
			disc, err := discovery.NewDiscoverer(fs, runner)
			if err == nil {
				discRes, diag := disc.Discover(cmd.Context(), targetDir)
				if diag == nil && discRes != nil {
					targetDir = discRes.RootDir
					if discRes.HasManifest {
						data, readErr := fs.ReadFile(discRes.ManifestPath)
						if readErr == nil {
							p := manifest.NewParser()
							parsed, _ := p.ParseStrict(discRes.ManifestPath, data)
							if parsed != nil {
								moduleName = parsed.Project.Module
							}
						}
					}
					if moduleName == "" && discRes.HasGoMod {
						data, readErr := fs.ReadFile(discRes.GoModPath)
						if readErr == nil {
							f, parseErr := modfile.Parse(discRes.GoModPath, data, nil)
							if parseErr == nil && f.Module != nil {
								moduleName = f.Module.Mod.Path
							}
						}
					}
				}
			}

			analyzer := architecture.NewAnalyzer(fs, architecture.AnalyzerConfig{
				ModuleName: moduleName,
				RootDir:    targetDir,
				Rules:      rules.DefaultRules(),
			})
			g, layers, violations, err := analyzer.BuildGraph(cmd.Context())
			if err != nil {
				return fmt.Errorf("building architecture graph: %w", err)
			}

			model := graph.NewGraphModel()
			for pkg, lyr := range layers {
				model.AddPackage(pkg, string(lyr))
			}
			if g != nil {
				allEdges := g.AllEdges()
				for _, edge := range allEdges {
					model.AddEdge(edge.From, edge.To)
				}

				edgeMap := make(map[string]map[int]graph.Edge)
				for _, edge := range allEdges {
					if _, ok := edgeMap[edge.File]; !ok {
						edgeMap[edge.File] = make(map[int]graph.Edge)
					}
					edgeMap[edge.File][edge.Line] = edge
				}

				isImportRule := func(ruleID string) bool {
					switch ruleID {
					case "ARCH-001", "ARCH-002", "ARCH-003", "ARCH-004", "ARCH-005",
						"ARCH-006", "ARCH-007", "ARCH-008", "ARCH-009", "ARCH-010":
						return true
					default:
						return false
					}
				}

				for _, v := range violations {
					if lines, ok := edgeMap[v.File]; ok {
						if edge, ok := lines[v.Line]; ok {
							model.AddViolation(edge.From, edge.To)
							continue
						}
					}
					// Only propagate to outgoing edges if the violation is an import/boundary rule
					if !isImportRule(v.RuleID) {
						continue
					}
					for _, node := range g.Nodes() {
						for _, f := range node.FilePaths {
							if f == v.File {
								for _, edge := range g.EdgesFrom(node.Path) {
									model.AddViolation(edge.From, edge.To)
								}
							}
						}
					}
				}
			}

			switch format {
			case "dot":
				graph.RenderDOT(cmd.OutOrStdout(), model, violationsOnly)
			case "mermaid":
				graph.RenderMermaid(cmd.OutOrStdout(), model, violationsOnly)
			case "ascii", "text", "":
				graph.RenderASCII(cmd.OutOrStdout(), model, violationsOnly)
			default:
				return fmt.Errorf("unsupported graph format %q: expected ascii, dot, or mermaid", format)
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&format, "format", "f", "ascii", "Graph output format (ascii, dot, mermaid)")
	cmd.Flags().BoolVar(&violationsOnly, "violations-only", false, "Render only packages involved in architectural violations")

	return cmd
}
