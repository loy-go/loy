package wiring_test

import (
	"context"
	"strings"
	"testing"

	"github.com/uloydev/loy/internal/filesystem"
	"github.com/uloydev/loy/internal/generator/builtin/wiring"
)

func TestWiringManager(t *testing.T) {
	memFS := filesystem.NewMemFileSystem()
	_ = memFS.MkdirAll("/app", 0755)

	mgr := wiring.NewSplicerManager(memFS)

	// Ensure creates default file
	err := mgr.EnsureWiringFile(context.Background(), "/app")
	if err != nil {
		t.Fatalf("ensure wiring failed: %v", err)
	}

	exists, _ := memFS.Exists("/app/internal/app/wiring.go")
	if !exists {
		t.Fatalf("expected wiring.go to be created")
	}

	data, _ := memFS.ReadFile("/app/internal/app/wiring.go")
	if !strings.Contains(string(data), "// loy:region:repositories") {
		t.Fatalf("missing region in created wiring.go")
	}

	// Calling again is idempotent
	err = mgr.EnsureWiringFile(context.Background(), "/app")
	if err != nil {
		t.Fatalf("idempotent ensure wiring failed: %v", err)
	}

	// Check artifacts produced
	arts := wiring.GenerateWiringArtifacts("invoice", "github.com/example/app")
	if len(arts) != 5 {
		t.Fatalf("expected 5 wiring artifacts, got %d", len(arts))
	}
}
