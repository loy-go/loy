package version_test

import (
	"runtime"
	"testing"

	"github.com/loy-go/loy/internal/version"
)

func TestGet(t *testing.T) {
	info := version.Get()

	if info.Version == "" {
		t.Error("expected non-empty Version")
	}
	if info.GoVersion != runtime.Version() {
		t.Errorf("expected GoVersion %q, got %q", runtime.Version(), info.GoVersion)
	}
	if info.Platform != runtime.GOOS+"/"+runtime.GOARCH {
		t.Errorf("expected Platform %q, got %q", runtime.GOOS+"/"+runtime.GOARCH, info.Platform)
	}
	if info.Compiler != runtime.Compiler {
		t.Errorf("expected Compiler %q, got %q", runtime.Compiler, info.Compiler)
	}
}
