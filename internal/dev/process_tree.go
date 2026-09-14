package dev

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"sync"
	"syscall"
	"time"
)

// ProcessTask models a supervised task to run in dev mode.
type ProcessTask struct {
	Name    string
	Command string
	Args    []string
	Dir     string
	Env     []string
}

// ManagedProcess coordinates the execution and termination of a child process group.
type ManagedProcess struct {
	task    ProcessTask
	logger  *PrefixedLogger
	cmd     *exec.Cmd
	mu      sync.Mutex
	running bool
	exited  chan struct{}
}

// NewManagedProcess creates a ManagedProcess.
func NewManagedProcess(task ProcessTask, logger *PrefixedLogger) *ManagedProcess {
	return &ManagedProcess{
		task:   task,
		logger: logger,
	}
}

// Start spawns the process in its own process group to prevent orphaned children.
func (p *ManagedProcess) Start(ctx context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.running {
		return fmt.Errorf("process %s is already running", p.task.Name)
	}

	cmd := exec.CommandContext(ctx, p.task.Command, p.task.Args...)
	cmd.Dir = p.task.Dir
	if len(p.task.Env) > 0 {
		cmd.Env = append(os.Environ(), p.task.Env...)
	}

	// Create new process group so child processes can be killed together
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}

	// Terminate the entire process group on context cancellation with graceful SIGTERM
	cmd.Cancel = func() error {
		if cmd.Process != nil && cmd.Process.Pid > 1 {
			return syscall.Kill(-cmd.Process.Pid, syscall.SIGTERM)
		}
		return nil
	}
	cmd.WaitDelay = 3 * time.Second

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("creating stdout pipe for %s: %w", p.task.Name, err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		_ = stdout.Close()
		return fmt.Errorf("creating stderr pipe for %s: %w", p.task.Name, err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("starting process %s: %w", p.task.Name, err)
	}

	p.cmd = cmd
	p.running = true
	p.exited = make(chan struct{})

	// Pipe output asynchronously
	go p.logger.PipeStream(p.task.Name, stdout)
	go p.logger.PipeStream(p.task.Name, stderr)

	// Monitor exit in background
	go func() {
		_ = cmd.Wait()
		p.mu.Lock()
		p.running = false
		close(p.exited)
		p.mu.Unlock()
	}()

	return nil
}

// Stop terminates the process group gracefully with SIGTERM, falling back to SIGKILL.
func (p *ManagedProcess) Stop(gracePeriod time.Duration) error {
	p.mu.Lock()
	if !p.running || p.cmd == nil || p.cmd.Process == nil {
		p.mu.Unlock()
		return nil
	}

	pid := p.cmd.Process.Pid
	if pid <= 1 {
		p.mu.Unlock()
		return nil
	}

	pgid, err := syscall.Getpgid(pid)
	if err != nil || pgid <= 1 {
		pgid = pid
	}
	exited := p.exited
	p.mu.Unlock()

	// Send SIGTERM to the process group (negative pgid)
	_ = syscall.Kill(-pgid, syscall.SIGTERM)

	select {
	case <-exited:
		return nil
	case <-time.After(gracePeriod):
	}

	// Forced SIGKILL fallback
	_ = syscall.Kill(-pgid, syscall.SIGKILL)
	select {
	case <-exited:
	case <-time.After(1 * time.Second):
	}
	return nil
}

// IsRunning reports whether the process is currently active.
func (p *ManagedProcess) IsRunning() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.running
}
