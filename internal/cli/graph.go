package cli

import (
	"archive/tar"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

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
		diffRef        string
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
			ctx := cmd.Context()
			opts := GetOptions(ctx)

			model, err := buildGraphModel(ctx, fs, runner, targetDir)
			if err != nil {
				return err
			}

			if diffRef != "" {
				if strings.HasPrefix(diffRef, "-") {
					return fmt.Errorf("invalid git ref %q: options not allowed", diffRef)
				}

				tmpBaseDir, err := os.MkdirTemp("", "loy-graph-diff-*")
				if err != nil {
					return fmt.Errorf("creating temp directory for diff base: %w", err)
				}
				defer func() { _ = os.RemoveAll(tmpBaseDir) }()

				res, err := runner.Run(ctx, targetDir, "git", "archive", "--format=tar", diffRef)
				if err != nil {
					return fmt.Errorf("extracting git ref %q: %w", diffRef, err)
				}
				if res.ExitCode != 0 {
					errMsg := strings.TrimSpace(string(res.Stderr))
					if errMsg == "" {
						errMsg = fmt.Sprintf("exit code %d", res.ExitCode)
					}
					return fmt.Errorf("extracting git ref %q: %s", diffRef, errMsg)
				}

				tarReader := tar.NewReader(bytes.NewReader(res.Stdout))
				for {
					hdr, err := tarReader.Next()
					if err == io.EOF {
						break
					}
					if err != nil {
						return fmt.Errorf("reading tar archive: %w", err)
					}
					targetPath, err := filesystem.CleanAndValidatePath(tmpBaseDir, filepath.Join(tmpBaseDir, hdr.Name))
					if err != nil {
						continue
					}
					switch hdr.Typeflag {
					case tar.TypeDir:
						_ = os.MkdirAll(targetPath, 0755)
					case tar.TypeReg:
						isRelevant := strings.HasSuffix(hdr.Name, ".go") ||
							filepath.Base(hdr.Name) == "go.mod" ||
							filepath.Base(hdr.Name) == "go.sum" ||
							filepath.Base(hdr.Name) == "loy.yaml"
						if !isRelevant {
							continue
						}
						_ = os.MkdirAll(filepath.Dir(targetPath), 0755)
						f, err := os.OpenFile(targetPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.FileMode(hdr.Mode))
						if err == nil {
							_, _ = io.Copy(f, tarReader)
							_ = f.Close()
						}
					}
				}

				baseModel, err := buildGraphModel(ctx, filesystem.NewOSFileSystem(), runner, tmpBaseDir)
				if err != nil {
					return fmt.Errorf("analyzing base git ref %q: %w", diffRef, err)
				}

				diff := graph.ComputeDiff(diffRef, baseModel, model)
				if opts.JSON {
					return json.NewEncoder(cmd.OutOrStdout()).Encode(diff)
				}

				switch format {
				case "mermaid":
					graph.RenderDiffMermaid(cmd.OutOrStdout(), diff)
				case "markdown", "md":
					graph.RenderDiffMarkdown(cmd.OutOrStdout(), diff)
				case "ascii", "text", "":
					graph.RenderDiffASCII(cmd.OutOrStdout(), diff)
				default:
					return fmt.Errorf("unsupported diff format %q: expected ascii, mermaid, or markdown", format)
				}
				return nil
			}

			if opts.JSON {
				return json.NewEncoder(cmd.OutOrStdout()).Encode(model)
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

	cmd.Flags().StringVarP(&format, "format", "f", "ascii", "Graph output format (ascii, dot, mermaid, markdown)")
	cmd.Flags().BoolVar(&violationsOnly, "violations-only", false, "Render only packages involved in architectural violations")
	cmd.Flags().StringVar(&diffRef, "diff", "", "Compare architectural dependencies against git ref (e.g. main, origin/main)")

	return cmd
}

func buildGraphModel(ctx context.Context, fs filesystem.FileSystem, runner process.Runner, targetDir string) (*graph.GraphModel, error) {
	moduleName := ""
	disc, err := discovery.NewDiscoverer(fs, runner)
	if err == nil {
		discRes, diag := disc.Discover(ctx, targetDir)
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
	g, layers, violations, err := analyzer.BuildGraph(ctx)
	if err != nil {
		return nil, fmt.Errorf("building architecture graph: %w", err)
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

	return model, nil
}
