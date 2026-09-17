package architecture

import (
	"context"
	"fmt"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"

	"github.com/loy-go/loy/internal/filesystem"
	"github.com/loy-go/loy/internal/graph"
	"golang.org/x/tools/go/packages"
)

// AnalyzerConfig tunes the architectural analysis.
type AnalyzerConfig struct {
	ModuleName string
	RootDir    string
	Deep       bool
	Strict     bool
	Rules      []Rule
	LayerMap   map[string]string
	Topology   *Topology
	Patterns   []PatternRule
}

// Analyzer orchestrates Phase 1 AST and Phase 2 Type analysis.
type Analyzer struct {
	fs  filesystem.FileSystem
	cfg AnalyzerConfig
}

// NewAnalyzer creates an architecture analyzer.
func NewAnalyzer(fs filesystem.FileSystem, cfg AnalyzerConfig) *Analyzer {
	return &Analyzer{
		fs:  fs,
		cfg: cfg,
	}
}

// BuildGraph executes discovery and returns the package Graph, Layer mappings, and unsuppressed violations.
func (a *Analyzer) BuildGraph(ctx context.Context) (*graph.Graph, map[string]Layer, []Violation, error) {
	analysis, layers, violations, err := a.analyze(ctx)
	if err != nil {
		return nil, nil, nil, err
	}
	return analysis.Graph, layers, violations, nil
}

// Run executes the two-phase analysis and returns all unsuppressed violations.
func (a *Analyzer) Run(ctx context.Context) ([]Violation, error) {
	_, _, violations, err := a.analyze(ctx)
	return violations, err
}

func (a *Analyzer) analyze(ctx context.Context) (*Analysis, map[string]Layer, []Violation, error) {
	classifier := NewClassifier(a.cfg.ModuleName, a.cfg.LayerMap)
	if len(a.cfg.Patterns) > 0 {
		classifier.SetPatterns(a.cfg.Patterns)
	}
	g := graph.New()
	layers := make(map[string]Layer)
	fset := token.NewFileSet()

	var files []*FileAST
	var allSuppressions []Suppression

	// Step 1: Discover and parse Go files using FileSystem abstraction
	walkFn := func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			base := filepath.Base(path)
			if base == ".git" || base == "vendor" || base == "node_modules" {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}

		data, readErr := a.fs.ReadFile(path)
		if readErr != nil {
			return fmt.Errorf("reading file %s: %w", path, readErr)
		}

		fileAST, parseErr := parser.ParseFile(fset, path, data, parser.ParseComments)
		if parseErr != nil {
			// In Phase 1 we record syntax issues or continue best-effort
			return nil
		}

		// Compute import path for this file
		relPath := path
		if rel, err := filepath.Rel(a.cfg.RootDir, path); err == nil && !strings.HasPrefix(rel, "..") {
			relPath = rel
		} else {
			evalRoot, errRoot := filepath.EvalSymlinks(a.cfg.RootDir)
			evalPath, errPath := filepath.EvalSymlinks(path)
			if errRoot == nil && errPath == nil {
				if r, errRel := filepath.Rel(evalRoot, evalPath); errRel == nil && !strings.HasPrefix(r, "..") {
					relPath = r
				}
			}
		}
		relPath = filepath.ToSlash(relPath)
		relPath = strings.TrimPrefix(relPath, "/")
		relPath = strings.TrimPrefix(relPath, "./")
		dir := filepath.ToSlash(filepath.Dir(relPath))
		pkgImport := a.cfg.ModuleName
		if dir != "." && dir != "" {
			pkgImport = a.cfg.ModuleName + "/" + dir
		}

		lyr := classifier.Classify(pkgImport)
		layers[pkgImport] = lyr

		// Register package node in graph
		g.AddNode(pkgImport, string(lyr), []string{path})

		// Register import edges
		for _, imp := range fileAST.Imports {
			target := strings.Trim(imp.Path.Value, `"`)
			pos := fset.Position(imp.Pos())
			g.AddEdge(pkgImport, target, path, pos.Line)
		}

		// Collect suppressions from comments
		for _, cg := range fileAST.Comments {
			for _, c := range cg.List {
				pos := fset.Position(c.Pos())
				if supp, ok := ParseSuppression(c.Text, path, pos.Line); ok {
					allSuppressions = append(allSuppressions, supp)
				}
			}
		}

		files = append(files, &FileAST{
			Path:     path,
			PkgPath:  pkgImport,
			Layer:    lyr,
			AST:      fileAST,
			FileSet:  fset,
			Comments: fileAST.Comments,
			Content:  data,
		})

		return nil
	}

	if err := a.fs.Walk(a.cfg.RootDir, walkFn); err != nil {
		return nil, nil, nil, fmt.Errorf("walking project root: %w", err)
	}

	topol := a.cfg.Topology
	if topol == nil {
		topol = NewDefaultTopology()
	}

	analysis := &Analysis{
		ModuleName:   a.cfg.ModuleName,
		Graph:        g,
		Files:        files,
		Classifier:   classifier,
		Suppressions: allSuppressions,
		IsDeep:       a.cfg.Deep,
		Topology:     topol,
	}

	// Step 2: Phase 2 Deep Analysis if requested
	if a.cfg.Deep {
		pkgCfg := &packages.Config{
			Context: ctx,
			Mode:    packages.NeedName | packages.NeedFiles | packages.NeedImports | packages.NeedTypes | packages.NeedSyntax | packages.NeedTypesInfo,
			Dir:     a.cfg.RootDir,
		}
		pkgs, loadErr := packages.Load(pkgCfg, "./...")
		if loadErr == nil {
			analysis.Packages = pkgs
		}
	}

	// Step 3: Run all configured rules
	var rawViolations []Violation
	for _, rule := range a.cfg.Rules {
		vs := rule.Check(ctx, analysis)
		rawViolations = append(rawViolations, vs...)
	}

	// Step 4: Apply suppression filtering
	finalViolations := FilterViolations(rawViolations, allSuppressions)
	return analysis, layers, finalViolations, nil
}
