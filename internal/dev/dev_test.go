package dev_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/loy-go/loy/internal/dev"
	"github.com/loy-go/loy/internal/filesystem"
)

func TestPrefixedLogger(t *testing.T) {
	buf := new(bytes.Buffer)
	logger := dev.NewPrefixedLogger(buf, true) // noColor

	logger.LogLine("api", "server listening on :8080")
	logger.LogLine("worker", "worker pool started")

	out := buf.String()
	if !strings.Contains(out, "[api] server listening on :8080") {
		t.Errorf("expected api prefix in output:\n%s", out)
	}
	if !strings.Contains(out, "[worker] worker pool started") {
		t.Errorf("expected worker prefix in output:\n%s", out)
	}
}

func TestDiscoverTasks(t *testing.T) {
	memFS := filesystem.NewMemFileSystem()

	// 1. Empty dir
	tasks := dev.DiscoverTasks("/app", memFS)
	if len(tasks) != 0 {
		t.Errorf("expected 0 tasks in empty dir, got %d", len(tasks))
	}

	// 2. cmd/api present
	_ = memFS.MkdirAll("/app/cmd/api", 0755)
	_ = memFS.WriteFile("/app/cmd/api/main.go", []byte("package main"), 0644)

	tasks = dev.DiscoverTasks("/app", memFS)
	if len(tasks) != 1 || tasks[0].Name != "api" {
		t.Errorf("expected task 'api', got %+v", tasks)
	}

	// 3. cmd/worker also present
	_ = memFS.MkdirAll("/app/cmd/worker", 0755)
	_ = memFS.WriteFile("/app/cmd/worker/main.go", []byte("package main"), 0644)

	tasks = dev.DiscoverTasks("/app", memFS)
	if len(tasks) != 2 {
		t.Errorf("expected 2 tasks (api, worker), got %d", len(tasks))
	}

	// 4. cmd/web present
	memFSWeb := filesystem.NewMemFileSystem()
	_ = memFSWeb.MkdirAll("/app/cmd/web", 0755)
	_ = memFSWeb.WriteFile("/app/cmd/web/main.go", []byte("package main"), 0644)

	tasks = dev.DiscoverTasks("/app", memFSWeb)
	if len(tasks) != 1 || tasks[0].Name != "web" {
		t.Errorf("expected task 'web', got %+v", tasks)
	}

	// 5. package.json with dev script and pnpm lockfile
	_ = memFSWeb.WriteFile("/app/package.json", []byte(`{"scripts":{"dev":"vite"}}`), 0644)
	_ = memFSWeb.WriteFile("/app/pnpm-lock.yaml", []byte(""), 0644)

	tasks = dev.DiscoverTasks("/app", memFSWeb)
	var hasVite bool
	for _, task := range tasks {
		if task.Name == "vite" {
			hasVite = true
			if task.Command != "pnpm" {
				t.Errorf("expected pnpm command for vite task, got %s", task.Command)
			}
		}
	}
	if !hasVite {
		t.Errorf("expected vite task from package.json with dev script")
	}
}

func TestResolvePackageManager(t *testing.T) {
	memFS := filesystem.NewMemFileSystem()

	if pm := dev.ResolvePackageManager("/test", memFS); pm != "npm" {
		t.Errorf("expected fallback to npm, got %s", pm)
	}

	_ = memFS.WriteFile("/test/yarn.lock", []byte(""), 0644)
	if pm := dev.ResolvePackageManager("/test", memFS); pm != "yarn" {
		t.Errorf("expected yarn, got %s", pm)
	}

	_ = memFS.WriteFile("/test/bun.lockb", []byte(""), 0644)
	if pm := dev.ResolvePackageManager("/test", memFS); pm != "bun" {
		t.Errorf("expected bun, got %s", pm)
	}

	_ = memFS.WriteFile("/test/pnpm-lock.yaml", []byte(""), 0644)
	if pm := dev.ResolvePackageManager("/test", memFS); pm != "pnpm" {
		t.Errorf("expected pnpm, got %s", pm)
	}
}

func TestWatcherDebounceAndClose(t *testing.T) {
	tempDir := t.TempDir()

	wOpts := dev.DefaultWatcherOptions(tempDir)
	wOpts.Debounce = 20 * time.Millisecond

	watcher, err := dev.NewWatcher(wOpts)
	if err != nil {
		t.Fatalf("creating watcher: %v", err)
	}
	defer func() { _ = watcher.Close() }()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	changeDetected := make(chan string, 10)

	go func() {
		_ = watcher.Watch(ctx, func(path string) {
			changeDetected <- path
		})
	}()

	// Allow fsnotify goroutine to start
	time.Sleep(100 * time.Millisecond)

	// Write a file to trigger change
	testFile := filepath.Join(tempDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("hello"), 0644); err != nil {
		t.Fatalf("writing test file: %v", err)
	}

	select {
	case <-changeDetected:
		// success
	case <-time.After(1 * time.Second):
		t.Errorf("timeout waiting for watcher change event")
	}
}
