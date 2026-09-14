package builtin_test

import (
	"context"
	"strings"
	"testing"

	"github.com/loy-go/loy/internal/generator"
	"github.com/loy-go/loy/internal/generator/builtin"
)

func TestK8sGenerator(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	mod := "github.com/example/demoapp"

	gen := builtin.NewK8sGenerator(mod).WithCapabilities(true, true, true)
	artifacts, err := gen.Generate(ctx, generator.Input{Name: "demoapp"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(artifacts) != 5 {
		t.Fatalf("expected 5 manifests, got %d", len(artifacts))
	}

	paths := make(map[string]string)
	for _, a := range artifacts {
		paths[a.Path] = string(a.Content)
	}

	// 1. Deployment
	dep, ok := paths["deploy/k8s/deployment.yaml"]
	if !ok {
		t.Fatalf("missing deploy/k8s/deployment.yaml")
	}
	if !strings.Contains(dep, "runAsNonRoot: true") {
		t.Errorf("missing runAsNonRoot in deployment:\n%s", dep)
	}
	if !strings.Contains(dep, "readOnlyRootFilesystem: true") {
		t.Errorf("missing readOnlyRootFilesystem in deployment:\n%s", dep)
	}
	if !strings.Contains(dep, "preStop:") {
		t.Errorf("missing preStop in deployment:\n%s", dep)
	}
	if !strings.Contains(dep, "path: /health/live") {
		t.Errorf("missing livenessProbe in deployment:\n%s", dep)
	}
	if !strings.Contains(dep, "path: /health/ready") {
		t.Errorf("missing readinessProbe in deployment:\n%s", dep)
	}
	if !strings.Contains(dep, "name: DATABASE_URL") {
		t.Errorf("missing DATABASE_URL secretKeyRef in deployment:\n%s", dep)
	}

	// 2. Service
	svc, ok := paths["deploy/k8s/service.yaml"]
	if !ok {
		t.Fatalf("missing deploy/k8s/service.yaml")
	}
	if !strings.Contains(svc, "port: 80") || !strings.Contains(svc, "targetPort: 8080") {
		t.Errorf("unexpected service ports:\n%s", svc)
	}

	// 3. ConfigMap
	cm, ok := paths["deploy/k8s/configmap.yaml"]
	if !ok {
		t.Fatalf("missing deploy/k8s/configmap.yaml")
	}
	if !strings.Contains(cm, "PORT: \"8080\"") {
		t.Errorf("missing PORT in configmap:\n%s", cm)
	}

	// 4. Ingress
	ing, ok := paths["deploy/k8s/ingress.yaml"]
	if !ok {
		t.Fatalf("missing deploy/k8s/ingress.yaml")
	}
	if !strings.Contains(ing, "demoapp.example.com") {
		t.Errorf("missing host in ingress:\n%s", ing)
	}

	// 5. Secret example
	sec, ok := paths["deploy/k8s/secret.yaml.example"]
	if !ok {
		t.Fatalf("missing deploy/k8s/secret.yaml.example")
	}
	if !strings.Contains(sec, "database-url:") {
		t.Errorf("missing database-url in secret example:\n%s", sec)
	}
}
