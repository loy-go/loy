package builtin

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/uloydev/loy/internal/generator"
	"github.com/uloydev/loy/internal/generator/model"
)

// K8sData holds parameters for Kubernetes manifest generation.
type K8sData struct {
	ModulePath   string
	ProjectName  string
	WithDatabase bool
	WithCache    bool
	WithQueue    bool
}

// K8sGenerator scaffolds standard cloud-native Kubernetes manifests in deploy/k8s/.
type K8sGenerator struct {
	modulePath   string
	withDatabase bool
	withCache    bool
	withQueue    bool
}

// NewK8sGenerator constructs K8sGenerator.
func NewK8sGenerator(modulePath string) *K8sGenerator {
	return &K8sGenerator{
		modulePath:   modulePath,
		withDatabase: true,
		withCache:    true,
		withQueue:    true,
	}
}

// WithCapabilities configures optional capability toggles.
func (g *K8sGenerator) WithCapabilities(db, cache, queue bool) *K8sGenerator {
	g.withDatabase = db
	g.withCache = cache
	g.withQueue = queue
	return g
}

func (g *K8sGenerator) Name() string {
	return "k8s"
}

func (g *K8sGenerator) Generate(ctx context.Context, input generator.Input) ([]model.Artifact, error) {
	projectName := filepath.Base(g.modulePath)
	if projectName == "." || projectName == "/" || projectName == "" {
		projectName = "app"
	}
	if input.Name != "" && input.Name != "k8s" {
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

	data := K8sData{
		ModulePath:   g.modulePath,
		ProjectName:  strings.ToLower(projectName),
		WithDatabase: g.withDatabase,
		WithCache:    g.withCache,
		WithQueue:    g.withQueue,
	}

	renderer := GetRenderer()

	manifests := []struct {
		tmplName string
		outPath  string
	}{
		{"deploy_k8s_deployment.yaml.tmpl", "deploy/k8s/deployment.yaml"},
		{"deploy_k8s_service.yaml.tmpl", "deploy/k8s/service.yaml"},
		{"deploy_k8s_configmap.yaml.tmpl", "deploy/k8s/configmap.yaml"},
		{"deploy_k8s_ingress.yaml.tmpl", "deploy/k8s/ingress.yaml"},
		{"deploy_k8s_secret.yaml.example.tmpl", "deploy/k8s/secret.yaml.example"},
	}

	var artifacts []model.Artifact
	for _, m := range manifests {
		tmplContent, err := ReadTemplate(m.tmplName)
		if err != nil {
			return nil, fmt.Errorf("reading %s: %w", m.tmplName, err)
		}
		rendered, err := renderer.Render(ctx, m.tmplName, tmplContent, data)
		if err != nil {
			return nil, fmt.Errorf("rendering %s: %w", m.outPath, err)
		}
		artifacts = append(artifacts, model.Artifact{
			Path:        m.outPath,
			Ownership:   model.DeveloperOwned,
			Permissions: 0644,
			Content:     rendered,
		})
	}

	return artifacts, nil
}
