package builtin

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/uloydev/loy/internal/generator"
	"github.com/uloydev/loy/internal/generator/model"
)

// DockerData holds parameters for Dockerfile and docker-compose generation.
type DockerData struct {
	ModulePath   string
	ProjectName  string
	Target       string
	WithDatabase bool
	WithCache    bool
	WithQueue    bool
	WithVite     bool
	IsWorkspace  bool
}

// DockerGenerator scaffolds production multi-stage Dockerfile and local docker-compose.yml.
type DockerGenerator struct {
	modulePath   string
	withDatabase bool
	withCache    bool
	withQueue    bool
	withVite     bool
	isWorkspace  bool
	target       string
}

// NewDockerGenerator constructs DockerGenerator.
func NewDockerGenerator(modulePath string) *DockerGenerator {
	return &DockerGenerator{
		modulePath:   modulePath,
		withDatabase: true,
		withCache:    true,
		withQueue:    true,
		target:       "api",
	}
}

// WithCapabilities sets database, cache, queue, vite toggles.
func (g *DockerGenerator) WithCapabilities(db, cache, queue, vite bool) *DockerGenerator {
	g.withDatabase = db
	g.withCache = cache
	g.withQueue = queue
	g.withVite = vite
	return g
}

// WithTarget sets the target binary to build (e.g. "api", "web").
func (g *DockerGenerator) WithTarget(target string) *DockerGenerator {
	g.target = target
	return g
}

// WithWorkspace toggles workspace root build context awareness.
func (g *DockerGenerator) WithWorkspace(isWs bool) *DockerGenerator {
	g.isWorkspace = isWs
	return g
}

func (g *DockerGenerator) Name() string {
	return "docker"
}

func (g *DockerGenerator) Generate(ctx context.Context, input generator.Input) ([]model.Artifact, error) {
	projectName := filepath.Base(g.modulePath)
	if projectName == "." || projectName == "/" || projectName == "" {
		projectName = "app"
	}
	if input.Name != "" && input.Name != "docker" {
		projectName = input.Name
	}

	target := g.target
	if input.Args != nil {
		if t, ok := input.Args["target"]; ok && t != "" {
			target = t
		}
		if v, ok := input.Args["vite"]; ok && (v == "true" || v == "1") {
			g.withVite = true
		}
		if db, ok := input.Args["database"]; ok {
			g.withDatabase = db == "true" || db == "1"
		}
		if c, ok := input.Args["cache"]; ok {
			g.withCache = c == "true" || c == "1"
		}
	}
	if target == "" {
		target = "api"
	}

	data := DockerData{
		ModulePath:   g.modulePath,
		ProjectName:  strings.ToLower(projectName),
		Target:       target,
		WithDatabase: g.withDatabase,
		WithCache:    g.withCache,
		WithQueue:    g.withQueue,
		WithVite:     g.withVite,
		IsWorkspace:  g.isWorkspace,
	}

	renderer := GetRenderer()

	// 1. Dockerfile
	dfTmpl, err := ReadTemplate("deploy_dockerfile.tmpl")
	if err != nil {
		return nil, fmt.Errorf("reading deploy_dockerfile template: %w", err)
	}
	dfContent, err := renderer.Render(ctx, "deploy_dockerfile", dfTmpl, data)
	if err != nil {
		return nil, fmt.Errorf("rendering Dockerfile: %w", err)
	}

	// 2. docker-compose.yml
	dcTmpl, err := ReadTemplate("deploy_compose.yml.tmpl")
	if err != nil {
		return nil, fmt.Errorf("reading deploy_compose template: %w", err)
	}
	dcContent, err := renderer.Render(ctx, "deploy_compose", dcTmpl, data)
	if err != nil {
		return nil, fmt.Errorf("rendering docker-compose.yml: %w", err)
	}

	return []model.Artifact{
		{
			Path:        "Dockerfile",
			Ownership:   model.DeveloperOwned,
			Permissions: 0644,
			Content:     dfContent,
		},
		{
			Path:        "docker-compose.yml",
			Ownership:   model.DeveloperOwned,
			Permissions: 0644,
			Content:     dcContent,
		},
	}, nil
}
