package builtin_test

import (
	"context"
	"strings"
	"testing"

	"github.com/uloydev/loy/internal/generator"
	"github.com/uloydev/loy/internal/generator/builtin"
)

func TestHelmGenerator(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	mod := "github.com/example/demoapp"

	gen := builtin.NewHelmGenerator(mod).WithCapabilities(true, true, true)
	artifacts, err := gen.Generate(ctx, generator.Input{Name: "demoapp"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(artifacts) != 8 {
		t.Fatalf("expected 8 helm chart files, got %d", len(artifacts))
	}

	files := make(map[string]string)
	for _, a := range artifacts {
		files[a.Path] = string(a.Content)
	}

	// 1. Chart.yaml
	chart, ok := files["deploy/helm/demoapp/Chart.yaml"]
	if !ok {
		t.Fatalf("missing deploy/helm/demoapp/Chart.yaml")
	}
	if !strings.Contains(chart, "apiVersion: v2") || !strings.Contains(chart, "name: demoapp") {
		t.Errorf("unexpected Chart.yaml content:\n%s", chart)
	}

	// 2. values.yaml
	vals, ok := files["deploy/helm/demoapp/values.yaml"]
	if !ok {
		t.Fatalf("missing deploy/helm/demoapp/values.yaml")
	}
	if !strings.Contains(vals, "replicaCount: 2") || !strings.Contains(vals, "runAsNonRoot: true") {
		t.Errorf("unexpected values.yaml content:\n%s", vals)
	}

	// 3. Deployment template
	dep, ok := files["deploy/helm/demoapp/templates/deployment.yaml"]
	if !ok {
		t.Fatalf("missing deploy/helm/demoapp/templates/deployment.yaml")
	}
	if !strings.Contains(dep, "path: /health/live") {
		t.Errorf("missing /health/live probe in helm deployment:\n%s", dep)
	}
}
