package builtin

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/loy-go/loy/internal/generator"
	"github.com/loy-go/loy/internal/generator/model"
)

// MetricsGenerator scaffolds Prometheus metrics recording, scrape endpoints, and Grafana dashboard definitions.
type MetricsGenerator struct {
	modulePath string
}

// NewMetricsGenerator constructs a MetricsGenerator.
func NewMetricsGenerator(modulePath string) *MetricsGenerator {
	return &MetricsGenerator{modulePath: modulePath}
}

func (g *MetricsGenerator) Name() string {
	return "metrics"
}

func (g *MetricsGenerator) Generate(ctx context.Context, input generator.Input) ([]model.Artifact, error) {
	name := input.Name
	if name == "" {
		name = filepath.Base(g.modulePath)
	}

	data := NewBaseData(name, g.modulePath, nil)
	renderer := GetRenderer()

	// 1. Platform metrics recorder
	metricsTmpl, err := ReadTemplate("platform_metrics.go.tmpl")
	if err != nil {
		return nil, fmt.Errorf("reading platform_metrics template: %w", err)
	}
	metricsRendered, err := renderer.RenderGo(ctx, "platform_metrics.go.tmpl", metricsTmpl, data)
	if err != nil {
		return nil, fmt.Errorf("rendering platform_metrics: %w", err)
	}

	// 2. Grafana Dashboard definition
	dashboardTmpl, err := ReadTemplate("deploy_grafana_dashboard.json.tmpl")
	if err != nil {
		return nil, fmt.Errorf("reading deploy_grafana_dashboard template: %w", err)
	}
	dashboardRendered, err := renderer.Render(ctx, "deploy_grafana_dashboard.json.tmpl", dashboardTmpl, data)
	if err != nil {
		return nil, fmt.Errorf("rendering deploy_grafana_dashboard: %w", err)
	}

	// 3. Prometheus Scraper config
	promTmpl, err := ReadTemplate("deploy_prometheus.yml.tmpl")
	if err != nil {
		return nil, fmt.Errorf("reading deploy_prometheus template: %w", err)
	}
	promRendered, err := renderer.Render(ctx, "deploy_prometheus.yml.tmpl", promTmpl, data)
	if err != nil {
		return nil, fmt.Errorf("rendering deploy_prometheus: %w", err)
	}

	return []model.Artifact{
		{
			Path:        "internal/platform/metrics/metrics.go",
			Content:     metricsRendered,
			Ownership:   model.DeveloperOwned,
			Permissions: 0644,
		},
		{
			Path:        "deploy/grafana/dashboard.json",
			Content:     dashboardRendered,
			Ownership:   model.DeveloperOwned,
			Permissions: 0644,
		},
		{
			Path:        "deploy/prometheus/prometheus.yml",
			Content:     promRendered,
			Ownership:   model.DeveloperOwned,
			Permissions: 0644,
		},
	}, nil
}
