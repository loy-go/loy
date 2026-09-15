package mcp

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

	"github.com/loy-go/loy/internal/architecture"
	"github.com/loy-go/loy/internal/architecture/rules"
	"github.com/loy-go/loy/internal/diagnostics"
	"github.com/loy-go/loy/internal/discovery"
	"github.com/loy-go/loy/internal/filesystem"
	"github.com/loy-go/loy/internal/generator"
	"github.com/loy-go/loy/internal/generator/builtin"
	"github.com/loy-go/loy/internal/generator/plan"
	"github.com/loy-go/loy/internal/graph"
	"github.com/loy-go/loy/internal/integration/providers/database"
	"github.com/loy-go/loy/internal/manifest"
	"github.com/loy-go/loy/internal/route"
	"golang.org/x/mod/modfile"
)

func (s *Server) registerDefaultToolsAndResources() {
	// 1. Tools: loy_check and loy_check_architecture
	checkToolDef := Tool{
		Name:        "loy_check_architecture",
		Description: "Verify Clean Architecture layer boundaries and check for rule violations (ARCH-001 through ARCH-015)",
		InputSchema: InputSchema{
			Type: "object",
			Properties: map[string]PropertyDef{
				"path": {
					Type:        "string",
					Description: "Target project directory (defaults to current project root)",
				},
				"deep": {
					Type:        "boolean",
					Description: "Enable deep type analysis using go/packages",
				},
			},
		},
	}
	s.RegisterTool(checkToolDef, s.handleLoyCheck)

	// Also register loy_check alias
	checkToolDefAlias := checkToolDef
	checkToolDefAlias.Name = "loy_check"
	s.RegisterTool(checkToolDefAlias, s.handleLoyCheck)

	// 2. Tool: loy_make_crud
	s.RegisterTool(Tool{
		Name:        "loy_make_crud",
		Description: "Scaffold a complete 4-layer Clean Architecture CRUD slice (Domain entity, Repository, Service, HTTP Handler, Tests) and wire it into the composition root",
		InputSchema: InputSchema{
			Type:     "object",
			Required: []string{"name"},
			Properties: map[string]PropertyDef{
				"name": {
					Type:        "string",
					Description: "Entity feature name (e.g. order, product, invoice)",
				},
				"fields": {
					Type:        "string",
					Description: "Space-separated entity fields (e.g. 'title:string:required price:float status:string')",
				},
				"modular": {
					Type:        "boolean",
					Description: "Generate isolated wire_<domain>.go composition root instead of flat wiring.go",
				},
			},
		},
	}, s.handleLoyMakeCRUD)

	// 3. Tool: loy_inspect_routes
	s.RegisterTool(Tool{
		Name:        "loy_inspect_routes",
		Description: "Inspect and list all registered HTTP and WebSocket route bindings in the application",
		InputSchema: InputSchema{
			Type: "object",
		},
	}, s.handleLoyInspectRoutes)

	// 4. Tool: loy_get_graph_diff
	s.RegisterTool(Tool{
		Name:        "loy_get_graph_diff",
		Description: "Compute architectural drift delta (added/removed packages, new/resolved violations) against a git base reference",
		InputSchema: InputSchema{
			Type: "object",
			Properties: map[string]PropertyDef{
				"base_ref": {
					Type:        "string",
					Description: "Git reference to compare against (e.g. main, HEAD~1, origin/main)",
				},
			},
		},
	}, s.handleLoyGetGraphDiff)

	// 5. Tool: loy_make_migration
	s.RegisterTool(Tool{
		Name:        "loy_make_migration",
		Description: "Create a new timestamped SQL database migration file powered by embedded Goose engine",
		InputSchema: InputSchema{
			Type:     "object",
			Required: []string{"name"},
			Properties: map[string]PropertyDef{
				"name": {
					Type:        "string",
					Description: "Descriptive migration name (e.g. create_orders_table)",
				},
			},
		},
	}, s.handleLoyMakeMigration)

	// Resources:
	// R1. loy://rules/catalog
	s.RegisterResource(Resource{
		URI:         "loy://rules/catalog",
		Name:        "Loy Clean Architecture Rules Catalog",
		Description: "Authoritative specification of Clean Architecture rules ARCH-001 through ARCH-015 with permitted vs prohibited patterns",
		MimeType:    "text/markdown",
	}, func(ctx context.Context, uri string) (*ResourceContent, error) {
		return &ResourceContent{
			URI:      uri,
			MimeType: "text/markdown",
			Text:     RulesCatalogMarkdown,
		}, nil
	})

	// R2. loy://project/manifest
	s.RegisterResource(Resource{
		URI:         "loy://project/manifest",
		Name:        "Loy Project Manifest",
		Description: "Current loy.yaml configuration and active capability matrix",
		MimeType:    "application/x-yaml",
	}, func(ctx context.Context, uri string) (*ResourceContent, error) {
		manifestPath := filepath.Join(s.projectRoot, "loy.yaml")
		data, err := s.fs.ReadFile(manifestPath)
		if err != nil {
			return nil, fmt.Errorf("reading loy.yaml: %w", err)
		}
		return &ResourceContent{
			URI:      uri,
			MimeType: "application/x-yaml",
			Text:     string(data),
		}, nil
	})

	// R3. loy://architecture/graph
	s.RegisterResource(Resource{
		URI:         "loy://architecture/graph",
		Name:        "Loy Architecture Dependency Graph",
		Description: "Mermaid dependency graph visualizing package layers and architectural boundaries",
		MimeType:    "text/vnd.mermaid",
	}, func(ctx context.Context, uri string) (*ResourceContent, error) {
		gm, err := s.buildGraphModel(ctx, s.projectRoot)
		if err != nil {
			return nil, fmt.Errorf("building graph model: %w", err)
		}
		var buf bytes.Buffer
		graph.RenderMermaid(&buf, gm, false)
		return &ResourceContent{
			URI:      uri,
			MimeType: "text/vnd.mermaid",
			Text:     buf.String(),
		}, nil
	})
}

func (s *Server) handleLoyCheck(ctx context.Context, args json.RawMessage) (*ToolCallResult, error) {
	var input struct {
		Path string `json:"path"`
		Deep bool   `json:"deep"`
	}
	if len(args) > 0 {
		_ = json.Unmarshal(args, &input)
	}

	targetDir := s.projectRoot
	if input.Path != "" {
		clean, err := filesystem.CleanAndValidatePath(s.projectRoot, input.Path)
		if err == nil {
			targetDir = clean
		}
	}

	moduleName := ""
	disc, err := discovery.NewDiscoverer(s.fs, s.runner)
	if err == nil {
		if discRes, diag := disc.Discover(ctx, targetDir); diag == nil && discRes != nil {
			targetDir = discRes.RootDir
			if discRes.HasGoMod {
				if data, err := s.fs.ReadFile(discRes.GoModPath); err == nil {
					if f, err := modfile.Parse(discRes.GoModPath, data, nil); err == nil && f.Module != nil {
						moduleName = f.Module.Mod.Path
					}
				}
			}
		}
	}

	analyzer := architecture.NewAnalyzer(s.fs, architecture.AnalyzerConfig{
		ModuleName: moduleName,
		RootDir:    targetDir,
		Rules:      rules.DefaultRules(),
		Deep:       input.Deep,
	})

	violations, err := analyzer.Run(ctx)
	if err != nil {
		return &ToolCallResult{
			Content: []ToolContent{{Type: "text", Text: fmt.Sprintf("Analysis error: %v", err)}},
			IsError: true,
		}, nil
	}

	if len(violations) == 0 {
		return &ToolCallResult{
			Content: []ToolContent{{
				Type: "text",
				Text: "✅ All architecture rules passed (0 violations across 4 layers).",
			}},
		}, nil
	}

	agentDiags := make([]diagnostics.AgentDiagnostic, 0, len(violations))
	for _, v := range violations {
		d := v.ToDiagnostic()
		agentDiags = append(agentDiags, diagnostics.ToAgentDiagnostic(d))
	}

	payload, _ := json.MarshalIndent(agentDiags, "", "  ")

	return &ToolCallResult{
		Content: []ToolContent{
			{
				Type: "text",
				Text: fmt.Sprintf("⚠️ %d Architectural Violation(s) Found:\n\n%s", len(violations), string(payload)),
			},
		},
		IsError: true,
	}, nil
}

func (s *Server) handleLoyMakeCRUD(ctx context.Context, args json.RawMessage) (*ToolCallResult, error) {
	var rawInput map[string]any
	if err := json.Unmarshal(args, &rawInput); err != nil {
		return nil, fmt.Errorf("parsing arguments: %w", err)
	}

	name, _ := rawInput["name"].(string)
	if name == "" {
		return nil, fmt.Errorf("entity name is required")
	}

	var fieldsStr string
	if fRaw, ok := rawInput["fields"]; ok {
		switch fVal := fRaw.(type) {
		case string:
			fieldsStr = fVal
		case []any:
			var parts []string
			for _, item := range fVal {
				if s, ok := item.(string); ok {
					parts = append(parts, s)
				}
			}
			fieldsStr = strings.Join(parts, " ")
		}
	}

	modular, _ := rawInput["modular"].(bool)

	moduleName := ""
	disc, err := discovery.NewDiscoverer(s.fs, s.runner)
	if err == nil {
		if discRes, diag := disc.Discover(ctx, s.projectRoot); diag == nil && discRes != nil {
			if discRes.HasGoMod {
				if data, err := s.fs.ReadFile(discRes.GoModPath); err == nil {
					if f, err := modfile.Parse(discRes.GoModPath, data, nil); err == nil && f.Module != nil {
						moduleName = f.Module.Mod.Path
					}
				}
			}
		}
	}

	gen := builtin.NewCRUDGenerator(moduleName)
	genInput := generator.Input{
		Name: name,
		Args: map[string]string{
			"fields":  fieldsStr,
			"modular": fmt.Sprintf("%t", modular),
		},
	}

	manifestPath := filepath.Join(s.projectRoot, "loy.yaml")
	if mData, err := s.fs.ReadFile(manifestPath); err == nil {
		p := manifest.NewParser()
		if m, diag := p.ParseStrict(manifestPath, mData); diag == nil && m != nil {
			if m.Defaults.HTTP != "" {
				genInput.Args["http"] = m.Defaults.HTTP
			}
			if m.Defaults.Database != "" {
				genInput.Args["database"] = m.Defaults.Database
			}
		}
	}

	artifacts, err := gen.Generate(ctx, genInput)
	if err != nil {
		return &ToolCallResult{
			Content: []ToolContent{{Type: "text", Text: fmt.Sprintf("Scaffolding error: %v", err)}},
			IsError: true,
		}, nil
	}

	builder := plan.NewBuilder(s.fs)
	p, err := builder.Build(ctx, s.projectRoot, artifacts, generator.Options{Force: false})
	if err != nil {
		return &ToolCallResult{
			Content: []ToolContent{{Type: "text", Text: fmt.Sprintf("Plan build error: %v", err)}},
			IsError: true,
		}, nil
	}

	executor := plan.NewExecutor(s.fs)
	if err := executor.Execute(ctx, p); err != nil {
		return &ToolCallResult{
			Content: []ToolContent{{Type: "text", Text: fmt.Sprintf("Execution error: %v", err)}},
			IsError: true,
		}, nil
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, "Successfully generated CRUD vertical slice for %q:\n", name)
	for _, op := range p.Operations {
		fmt.Fprintf(&sb, "  + %s (%s)\n", op.Path, op.Type)
	}

	return &ToolCallResult{
		Content: []ToolContent{{Type: "text", Text: sb.String()}},
	}, nil
}

func (s *Server) handleLoyInspectRoutes(ctx context.Context, _ json.RawMessage) (*ToolCallResult, error) {
	routes, err := route.InspectRoutes(s.fs, s.projectRoot)
	if err != nil {
		return &ToolCallResult{
			Content: []ToolContent{{Type: "text", Text: fmt.Sprintf("Error inspecting routes: %v", err)}},
			IsError: true,
		}, nil
	}

	if len(routes) == 0 {
		return &ToolCallResult{
			Content: []ToolContent{{Type: "text", Text: "No registered HTTP or WebSocket route bindings discovered."}},
		}, nil
	}

	payload, err := json.MarshalIndent(routes, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("formatting routes: %w", err)
	}

	return &ToolCallResult{
		Content: []ToolContent{{Type: "text", Text: string(payload)}},
	}, nil
}

func (s *Server) handleLoyGetGraphDiff(ctx context.Context, args json.RawMessage) (*ToolCallResult, error) {
	var input struct {
		BaseRef string `json:"base_ref"`
	}
	if len(args) > 0 {
		_ = json.Unmarshal(args, &input)
	}

	baseRef := input.BaseRef
	if baseRef == "" {
		baseRef = "main"
	}

	currModel, err := s.buildGraphModel(ctx, s.projectRoot)
	if err != nil {
		return nil, fmt.Errorf("analyzing current architecture: %w", err)
	}

	baseModel, err := s.buildBaseGraphModel(ctx, baseRef)
	if err != nil {
		// If base ref retrieval fails, fallback to comparing against empty model
		baseModel = graph.NewGraphModel()
	}

	diff := graph.ComputeDiff(baseRef, baseModel, currModel)
	var buf bytes.Buffer
	graph.RenderDiffMarkdown(&buf, diff)

	return &ToolCallResult{
		Content: []ToolContent{{Type: "text", Text: buf.String()}},
	}, nil
}

func (s *Server) handleLoyMakeMigration(ctx context.Context, args json.RawMessage) (*ToolCallResult, error) {
	var input struct {
		Name string `json:"name"`
	}
	if err := json.Unmarshal(args, &input); err != nil {
		return nil, fmt.Errorf("parsing arguments: %w", err)
	}

	if input.Name == "" {
		return nil, fmt.Errorf("migration name is required")
	}

	migrationsDir := "migrations"
	manifestPath := filepath.Join(s.projectRoot, "loy.yaml")
	if mData, err := s.fs.ReadFile(manifestPath); err == nil {
		p := manifest.NewParser()
		if m, diag := p.ParseStrict(manifestPath, mData); diag == nil && m != nil {
			if dbInt, ok := m.Integrations["database"]; ok {
				if dVal, ok := dbInt.Config["migrations_dir"].(string); ok && dVal != "" {
					migrationsDir = dVal
				}
			}
		}
	}

	gRunner := database.NewGooseRunner(s.fs, nil)
	filePath, err := gRunner.Create(ctx, migrationsDir, input.Name)
	if err != nil {
		return &ToolCallResult{
			Content: []ToolContent{{Type: "text", Text: fmt.Sprintf("Creating migration error: %v", err)}},
			IsError: true,
		}, nil
	}

	return &ToolCallResult{
		Content: []ToolContent{{Type: "text", Text: fmt.Sprintf("Created migration file: %s", filePath)}},
	}, nil
}

func (s *Server) buildGraphModel(ctx context.Context, targetDir string) (*graph.GraphModel, error) {
	moduleName := ""
	disc, err := discovery.NewDiscoverer(s.fs, s.runner)
	if err == nil {
		discRes, diag := disc.Discover(ctx, targetDir)
		if diag == nil && discRes != nil {
			targetDir = discRes.RootDir
			if discRes.HasGoMod {
				if data, err := s.fs.ReadFile(discRes.GoModPath); err == nil {
					if f, err := modfile.Parse(discRes.GoModPath, data, nil); err == nil && f.Module != nil {
						moduleName = f.Module.Mod.Path
					}
				}
			}
		}
	}

	analyzer := architecture.NewAnalyzer(s.fs, architecture.AnalyzerConfig{
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

		for _, v := range violations {
			if lines, ok := edgeMap[v.File]; ok {
				if edge, ok := lines[v.Line]; ok {
					model.AddViolation(edge.From, edge.To)
					continue
				}
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

func (s *Server) buildBaseGraphModel(ctx context.Context, baseRef string) (*graph.GraphModel, error) {
	if s.runner == nil || strings.HasPrefix(baseRef, "-") {
		return graph.NewGraphModel(), nil
	}

	tmpBaseDir, err := os.MkdirTemp("", "loy-mcp-diff-*")
	if err != nil {
		return graph.NewGraphModel(), err
	}
	defer func() { _ = os.RemoveAll(tmpBaseDir) }()

	res, err := s.runner.Run(ctx, s.projectRoot, "git", "archive", "--format=tar", baseRef)
	if err != nil || res.ExitCode != 0 {
		return graph.NewGraphModel(), nil
	}

	tarReader := tar.NewReader(bytes.NewReader(res.Stdout))
	for {
		hdr, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			break
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

	return s.buildGraphModel(ctx, tmpBaseDir)
}
