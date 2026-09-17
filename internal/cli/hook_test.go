package cli

import (
	"bytes"
	"testing"

	"github.com/loy-go/loy/internal/filesystem"
)

func TestHookInstall(t *testing.T) {
	memFS := filesystem.NewMemFileSystem()
	_ = memFS.MkdirAll("/workspace/.git", 0755)

	cmd := NewRootCmdWithFS(memFS, nil)
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"hook", "install", "/workspace"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	exists, err := memFS.Exists("/workspace/.git/hooks/pre-commit")
	if err != nil || !exists {
		t.Fatalf("expected pre-commit hook to exist, err: %v", err)
	}

	content, err := memFS.ReadFile("/workspace/.git/hooks/pre-commit")
	if err != nil {
		t.Fatalf("failed reading hook content: %v", err)
	}

	if !bytes.Contains(content, []byte("loy check --quiet")) {
		t.Errorf("expected hook to contain 'loy check --quiet', got:\n%s", string(content))
	}

	pushExists, err := memFS.Exists("/workspace/.git/hooks/pre-push")
	if err != nil || !pushExists {
		t.Fatalf("expected pre-push hook to exist, err: %v", err)
	}

	pushContent, err := memFS.ReadFile("/workspace/.git/hooks/pre-push")
	if err != nil {
		t.Fatalf("failed reading pre-push hook content: %v", err)
	}

	if !bytes.Contains(pushContent, []byte("go test -short")) {
		t.Errorf("expected hook to contain 'go test -short', got:\n%s", string(pushContent))
	}
}
