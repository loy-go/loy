package builtin

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/loy-go/loy/internal/generator"
	"github.com/loy-go/loy/internal/generator/model"
)

// HelmData holds parameters for Helm chart generation.
type HelmData struct {
	ModulePath   string
	ProjectName  string
	WithDatabase bool
	WithCache    bool
	WithQueue    bool
}

// HelmGenerator scaffolds standard Helm charts in deploy/helm/<app>/.
type HelmGenerator struct {
	modulePath   string
	withDatabase bool
	withCache    bool
	withQueue    bool
}

// NewHelmGenerator constructs HelmGenerator.
func NewHelmGenerator(modulePath string) *HelmGenerator {
	return &HelmGenerator{
		modulePath:   modulePath,
		withDatabase: true,
		withCache:    true,
		withQueue:    true,
	}
}

// WithCapabilities configures optional capability toggles.
func (g *HelmGenerator) WithCapabilities(db, cache, queue bool) *HelmGenerator {
	g.withDatabase = db
	g.withCache = cache
	g.withQueue = queue
	return g
}

func (g *HelmGenerator) Name() string {
	return "helm"
}

func (g *HelmGenerator) Generate(ctx context.Context, input generator.Input) ([]model.Artifact, error) {
	projectName := filepath.Base(g.modulePath)
	if projectName == "." || projectName == "/" || projectName == "" {
		projectName = "app"
	}
	if input.Name != "" && input.Name != "helm" {
		projectName = input.Name
	}

	if input.Args != nil {
		if db, ok := input.Args["database"]; ok {
			g.withDatabase = db == "true" || db == "1"
		}
		if c, ok := input.Args["cache"]; ok {
			g.withCache = c == "true" || c == "1"
		}
	}

	cleanProject := strings.ToLower(projectName)
	data := HelmData{
		ModulePath:   g.modulePath,
		ProjectName:  cleanProject,
		WithDatabase: g.withDatabase,
		WithCache:    g.withCache,
		WithQueue:    g.withQueue,
	}

	renderer := GetRenderer()

	chartDir := filepath.ToSlash(filepath.Join("deploy/helm", cleanProject))
	templates := []struct {
		tmplName string
		outRel   string
	}{
		{"deploy_helm_chart.yaml.tmpl", "Chart.yaml"},
		{"deploy_helm_values.yaml.tmpl", "values.yaml"},
		{"deploy_helm_helpers.tpl.tmpl", "templates/_helpers.tpl"},
		{"deploy_helm_deployment.yaml.tmpl", "templates/deployment.yaml"},
		{"deploy_helm_service.yaml.tmpl", "templates/service.yaml"},
		{"deploy_helm_configmap.yaml.tmpl", "templates/configmap.yaml"},
		{"deploy_helm_ingress.yaml.tmpl", "templates/ingress.yaml"},
		{"deploy_helm_hpa.yaml.tmpl", "templates/hpa.yaml"},
	}

	var artifacts []model.Artifact
	for _, t := range templates {
		tmplContent, err := ReadTemplate(t.tmplName)
		if err != nil {
			return nil, fmt.Errorf("reading %s: %w", t.tmplName, err)
		}
		rendered, err := renderer.Render(ctx, t.tmplName, tmplContent, data)
		if err != nil {
			return nil, fmt.Errorf("rendering %s: %w", t.outRel, err)
		}
		outPath := filepath.ToSlash(filepath.Join(chartDir, t.outRel))
		artifacts = append(artifacts, model.Artifact{
			Path:        outPath,
			Ownership:   model.DeveloperOwned,
			Permissions: 0644,
			Content:     rendered,
		})
	}

	return artifacts, nil
}
