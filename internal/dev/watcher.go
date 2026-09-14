package dev

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

// WatcherOptions configures directory watching rules.
type WatcherOptions struct {
	RootDir     string
	Debounce    time.Duration
	ExcludeDirs []string
}

// DefaultWatcherOptions returns standard sensible dev watch options.
func DefaultWatcherOptions(rootDir string) WatcherOptions {
	return WatcherOptions{
		RootDir:  rootDir,
		Debounce: 200 * time.Millisecond,
		ExcludeDirs: []string{
			".git",
			"node_modules",
			"tmp",
			"dist",
			"bin",
			".idea",
			".vscode",
		},
	}
}

// Watcher listens for file change events under a directory tree with debounce filtering.
type Watcher struct {
	opts    WatcherOptions
	watcher *fsnotify.Watcher
	mu      sync.Mutex
	closed  bool
}

// NewWatcher constructs a Watcher.
func NewWatcher(opts WatcherOptions) (*Watcher, error) {
	if opts.Debounce <= 0 {
		opts.Debounce = 200 * time.Millisecond
	}

	fw, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("creating fsnotify watcher: %w", err)
	}

	w := &Watcher{
		opts:    opts,
		watcher: fw,
	}

	// Add root directory and all subdirectories recursively
	if err := w.addRecursive(opts.RootDir); err != nil {
		_ = fw.Close()
		return nil, err
	}

	return w, nil
}

func (w *Watcher) isExcluded(path string) bool {
	// Relative path from RootDir to avoid matching the root directory path segments
	rel, err := filepath.Rel(w.opts.RootDir, path)
	if err != nil {
		rel = path
	}
	clean := filepath.ToSlash(rel)
	parts := strings.Split(clean, "/")
	for _, part := range parts {
		if part == "" || part == "." {
			continue
		}
		for _, ex := range w.opts.ExcludeDirs {
			if part == ex {
				return true
			}
		}
	}
	return false
}

func (w *Watcher) addRecursive(root string) error {
	return filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil // skip unreadable items
		}
		if d.IsDir() {
			if w.isExcluded(path) {
				return filepath.SkipDir
			}
			if err := w.watcher.Add(path); err != nil {
				return nil
			}
		}
		return nil
	})
}

// Watch blocks listening for changes and yields debounced change notifications on onChange.
func (w *Watcher) Watch(ctx context.Context, onChange func(path string)) error {
	var (
		timer    *time.Timer
		lastPath string
		mu       sync.Mutex
	)

	defer func() {
		mu.Lock()
		if timer != nil {
			timer.Stop()
			timer = nil
		}
		mu.Unlock()
	}()

	for {
		select {
		case <-ctx.Done():
			_ = w.Close()
			return ctx.Err()

		case event, ok := <-w.watcher.Events:
			if !ok {
				return nil
			}

			// Ignore chmod events and excluded files
			if event.Op&fsnotify.Chmod == fsnotify.Chmod {
				continue
			}
			if w.isExcluded(event.Name) {
				continue
			}

			// Add newly created directories dynamically
			if event.Op&fsnotify.Create == fsnotify.Create {
				if fi, err := os.Stat(event.Name); err == nil && fi.IsDir() {
					_ = w.addRecursive(event.Name)
				}
			}

			mu.Lock()
			lastPath = event.Name
			if timer == nil {
				timer = time.AfterFunc(w.opts.Debounce, func() {
					mu.Lock()
					p := lastPath
					timer = nil
					mu.Unlock()
					if p != "" && ctx.Err() == nil {
						onChange(p)
					}
				})
			} else {
				timer.Reset(w.opts.Debounce)
			}
			mu.Unlock()

		case err, ok := <-w.watcher.Errors:
			if !ok {
				return nil
			}
			// Non-fatal watch error
			_ = err
		}
	}
}

// Close closes the underlying fsnotify watcher.
func (w *Watcher) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.closed {
		return nil
	}
	w.closed = true
	return w.watcher.Close()
}
