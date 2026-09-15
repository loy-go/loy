package builtin

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/loy-go/loy/internal/generator"
	"github.com/loy-go/loy/internal/generator/model"
)

// CIData holds parameters for CI pipeline generation.
type CIData struct {
	ModulePath      string
	ProjectName     string
	Provider        string // "github" or "gitlab"
	DatabaseDialect string // "postgres", "sqlite", "mysql"
	WithDatabase    bool
	WithCache       bool
	WithQueue       bool
}

// CIGenerator scaffolds continuous integration workflows for GitHub Actions or GitLab CI.
type CIGenerator struct {
	modulePath      string
	provider        string
	databaseDialect string
	withDatabase    bool
	withCache       bool
	withQueue       bool
}

// NewCIGenerator constructs CIGenerator.
func NewCIGenerator(modulePath string) *CIGenerator {
	return &CIGenerator{
		modulePath:      modulePath,
		provider:        "github",
		databaseDialect: "postgres",
		withDatabase:    true,
		withCache:       true,
		withQueue:       true,
	}
}

// WithProvider sets the CI provider ("github" or "gitlab").
func (g *CIGenerator) WithProvider(provider string) *CIGenerator {
	if provider != "" {
		g.provider = strings.ToLower(provider)
	}
	return g
}

// WithDatabaseDialect sets the database dialect ("postgres", "sqlite", "mysql", "none").
func (g *CIGenerator) WithDatabaseDialect(dialect string) *CIGenerator {
	if dialect != "" {
		g.databaseDialect = dialect
		if dialect == "none" {
			g.withDatabase = false
		}
	}
	return g
}

// WithCapabilities configures optional capability toggles.
func (g *CIGenerator) WithCapabilities(db, cache, queue bool) *CIGenerator {
	g.withDatabase = db
	g.withCache = cache
	g.withQueue = queue
	return g
}

func (g *CIGenerator) Name() string {
	return "ci"
}

func (g *CIGenerator) Generate(ctx context.Context, input generator.Input) ([]model.Artifact, error) {
	projectName := filepath.Base(g.modulePath)
	if projectName == "." || projectName == "/" || projectName == "" {
		projectName = "app"
	}
	if input.Name != "" && input.Name != "ci" {
		projectName = input.Name
	}

	provider := g.provider
	if input.Args != nil {
		if p, ok := input.Args["provider"]; ok && p != "" {
			provider = strings.ToLower(p)
		}
		if db, ok := input.Args["database"]; ok {
			g.databaseDialect = db
			g.withDatabase = db != "none" && db != "" && db != "false"
		}
		if c, ok := input.Args["cache"]; ok {
			g.withCache = c == "true" || c == "1"
		}
	}

	data := CIData{
		ModulePath:      g.modulePath,
		ProjectName:     strings.ToLower(projectName),
		Provider:        provider,
		DatabaseDialect: g.databaseDialect,
		WithDatabase:    g.withDatabase,
		WithCache:       g.withCache,
		WithQueue:       g.withQueue,
	}

	renderer := GetRenderer()

	var tmplName, outPath string
	if provider == "gitlab" {
		tmplName = "deploy_ci_gitlab.yml.tmpl"
		outPath = ".gitlab-ci.yml"
	} else {
		tmplName = "deploy_ci_github.yml.tmpl"
		outPath = ".github/workflows/ci.yml"
	}

	tmplContent, err := ReadTemplate(tmplName)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", tmplName, err)
	}

	rendered, err := renderer.Render(ctx, tmplName, tmplContent, data)
	if err != nil {
		return nil, fmt.Errorf("rendering %s: %w", outPath, err)
	}

	return []model.Artifact{
		{
			Path:        outPath,
			Ownership:   model.DeveloperOwned,
			Permissions: 0644,
			Content:     rendered,
		},
	}, nil
}
