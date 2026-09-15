package preset_test

import (
	"testing"

	"github.com/loy-go/loy/internal/manifest"
	"github.com/loy-go/loy/internal/preset"
)

func TestRegistry_Presets(t *testing.T) {
	reg := preset.NewRegistry()
	names := reg.Names()
	if len(names) != 6 {
		t.Fatalf("expected 6 presets, got %d", len(names))
	}

	parser := manifest.NewParser()
	val := manifest.NewValidator()

	for _, name := range names {
		p, ok := reg.Get(name)
		if !ok {
			t.Fatalf("preset %s not found in registry", name)
		}
		if p.Description == "" {
			t.Errorf("expected non-empty description for preset %s", name)
		}
		mObj := p.Manifest("demo")
		if mObj == nil || mObj.Project.Name != "demo" {
			t.Errorf("preset %s returned invalid manifest object: %+v", name, mObj)
		}
		yamlStr := p.MaterializeYAML("demo")
		m, diag := parser.ParseStrict("loy.yaml", []byte(yamlStr))
		if diag != nil {
			t.Fatalf("preset %s produced invalid YAML: %+v", name, diag)
		}
		diags := val.Validate("loy.yaml", m)
		if len(diags) > 0 {
			t.Fatalf("preset %s failed validation: %+v", name, diags)
		}
	}
}
