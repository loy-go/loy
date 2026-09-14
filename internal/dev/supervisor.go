package dev

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"

	"github.com/loy-go/loy/internal/filesystem"
)

// Supervisor coordinates multiple ManagedProcesses and triggers hot-reloads on file change.
type Supervisor struct {
	rootDir     string
	tasks       []ProcessTask
	logger      *PrefixedLogger
	watcher     *Watcher
	processes   []*ManagedProcess
	gracePeriod time.Duration
	mu          sync.Mutex
	stopped     bool
}

// SupervisorOptions holds initialization settings.
type SupervisorOptions struct {
	RootDir     string
	Tasks       []ProcessTask
	Debounce    time.Duration
	GracePeriod time.Duration
	Stdout      io.Writer
	NoColor     bool
}

// NewSupervisor constructs a new Supervisor.
func NewSupervisor(opts SupervisorOptions, fs filesystem.FileSystem) (*Supervisor, error) {
	if fs == nil {
		fs = filesystem.NewOSFileSystem()
	}
	if opts.Stdout == nil {
		opts.Stdout = os.Stdout
	}
	if opts.GracePeriod <= 0 {
		opts.GracePeriod = 3 * time.Second
	}
	if opts.Debounce <= 0 {
		opts.Debounce = 200 * time.Millisecond
	}

	logger := NewPrefixedLogger(opts.Stdout, opts.NoColor)

	wOpts := DefaultWatcherOptions(opts.RootDir)
	wOpts.Debounce = opts.Debounce

	watcher, err := NewWatcher(wOpts)
	if err != nil {
		return nil, fmt.Errorf("initializing file watcher: %w", err)
	}

	// Graceful warning if views/ directory exists but templ CLI is not found in PATH
	viewsPath := filepath.Join(opts.RootDir, "views")
	if exists, _ := fs.Exists(viewsPath); exists {
		if _, err := exec.LookPath("templ"); err != nil {
			logger.LogLine("supervisor", "Warning: 'views/' directory detected but 'templ' CLI is not found in PATH.")
			logger.LogLine("supervisor", "To enable live Templ compilation, install it with: go install github.com/a-h/templ/cmd/templ@latest")
		}
	}

	s := &Supervisor{
		rootDir:     opts.RootDir,
		tasks:       opts.Tasks,
		logger:      logger,
		watcher:     watcher,
		gracePeriod: opts.GracePeriod,
	}

	return s, nil
}

// ResolvePackageManager determines the JS package manager based on lockfiles.
func ResolvePackageManager(rootDir string, fs filesystem.FileSystem) string {
	if fs == nil {
		fs = filesystem.NewOSFileSystem()
	}
	if exists, _ := fs.Exists(filepath.Join(rootDir, "pnpm-lock.yaml")); exists {
		return "pnpm"
	}
	if exists, _ := fs.Exists(filepath.Join(rootDir, "bun.lockb")); exists {
		return "bun"
	}
	if exists, _ := fs.Exists(filepath.Join(rootDir, "yarn.lock")); exists {
		return "yarn"
	}
	return "npm"
}

// DiscoverTasks inspects the project directory and returns standard default process tasks.
func DiscoverTasks(rootDir string, fs filesystem.FileSystem) []ProcessTask {
	if fs == nil {
		fs = filesystem.NewOSFileSystem()
	}

	var tasks []ProcessTask

	// 1. Check for cmd/web
	webPath := filepath.Join(rootDir, "cmd", "web")
	if exists, _ := fs.Exists(webPath); exists {
		tasks = append(tasks, ProcessTask{
			Name:    "web",
			Command: "go",
			Args:    []string{"run", "./cmd/web"},
			Dir:     rootDir,
		})
	}

	// 2. Check for cmd/api
	apiPath := filepath.Join(rootDir, "cmd", "api")
	if exists, _ := fs.Exists(apiPath); exists {
		tasks = append(tasks, ProcessTask{
			Name:    "api",
			Command: "go",
			Args:    []string{"run", "./cmd/api"},
			Dir:     rootDir,
		})
	} else if len(tasks) == 0 {
		// Fallback to main.go in root if neither cmd/web nor cmd/api exists
		mainPath := filepath.Join(rootDir, "main.go")
		if exists, _ := fs.Exists(mainPath); exists {
			tasks = append(tasks, ProcessTask{
				Name:    "app",
				Command: "go",
				Args:    []string{"run", "."},
				Dir:     rootDir,
			})
		}
	}

	// 3. Check for cmd/worker
	workerPath := filepath.Join(rootDir, "cmd", "worker")
	if exists, _ := fs.Exists(workerPath); exists {
		tasks = append(tasks, ProcessTask{
			Name:    "worker",
			Command: "go",
			Args:    []string{"run", "./cmd/worker"},
			Dir:     rootDir,
		})
	}

	// 4. Check for Templ in views/
	viewsPath := filepath.Join(rootDir, "views")
	if exists, _ := fs.Exists(viewsPath); exists {
		if _, err := exec.LookPath("templ"); err == nil {
			tasks = append(tasks, ProcessTask{
				Name:    "templ",
				Command: "templ",
				Args:    []string{"generate", "--watch"},
				Dir:     rootDir,
			})
		}
	}

	// 5. Check for package.json with dev script (Vite)
	pkgPath := filepath.Join(rootDir, "package.json")
	if exists, _ := fs.Exists(pkgPath); exists {
		pkgData, err := fs.ReadFile(pkgPath)
		if err == nil {
			var pkgInfo struct {
				Scripts map[string]string `json:"scripts"`
			}
			if json.Unmarshal(pkgData, &pkgInfo) == nil {
				if _, hasDev := pkgInfo.Scripts["dev"]; hasDev {
					pm := ResolvePackageManager(rootDir, fs)
					tasks = append(tasks, ProcessTask{
						Name:    "vite",
						Command: pm,
						Args:    []string{"run", "dev"},
						Dir:     rootDir,
					})
				}
			}
		}
	}

	return tasks
}

// Start spawns all configured processes.
func (s *Supervisor) startProcesses(ctx context.Context) error {
	s.processes = nil
	for _, task := range s.tasks {
		mp := NewManagedProcess(task, s.logger)
		s.logger.LogLine("supervisor", fmt.Sprintf("starting %s...", task.Name))
		if err := mp.Start(ctx); err != nil {
			s.logger.LogLine("supervisor", fmt.Sprintf("failed to start %s: %v", task.Name, err))
		}
		s.processes = append(s.processes, mp)
	}
	return nil
}

// stopProcesses stops all running processes.
func (s *Supervisor) stopProcesses() {
	var wg sync.WaitGroup
	for _, mp := range s.processes {
		if mp.IsRunning() {
			wg.Add(1)
			go func(p *ManagedProcess) {
				defer wg.Done()
				_ = p.Stop(s.gracePeriod)
			}(mp)
		}
	}
	wg.Wait()
}

// Run starts the supervisor, listens for file changes, and manages live restarts until context cancellation.
func (s *Supervisor) Run(ctx context.Context) error {
	s.mu.Lock()
	if err := s.startProcesses(ctx); err != nil {
		s.mu.Unlock()
		return err
	}
	s.mu.Unlock()

	defer func() {
		s.mu.Lock()
		s.stopped = true
		s.logger.LogLine("supervisor", "shutting down processes...")
		s.stopProcesses()
		_ = s.watcher.Close()
		s.mu.Unlock()
	}()

	return s.watcher.Watch(ctx, func(changedPath string) {
		s.mu.Lock()
		defer s.mu.Unlock()

		if s.stopped || ctx.Err() != nil {
			return
		}

		rel, err := filepath.Rel(s.rootDir, changedPath)
		if err != nil {
			rel = changedPath
		}
		s.logger.LogLine("supervisor", fmt.Sprintf("change detected in %s, restarting...", rel))

		s.stopProcesses()
		if !s.stopped && ctx.Err() == nil {
			_ = s.startProcesses(ctx)
		}
	})
}
