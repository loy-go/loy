package dev

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"
)

// ProcessStatusSnapshot captures the state of a supervised task.
type ProcessStatusSnapshot struct {
	Name     string
	PID      int
	Running  bool
	Restarts int
}

// TUIOptions configures the interactive dashboard renderer.
type TUIOptions struct {
	ProjectName string
	Stdout      io.Writer
	Stdin       io.Reader
	RefreshRate time.Duration
	NoColor     bool
}

// TUIDashboard renders a live terminal interface for the dev supervisor.
type TUIDashboard struct {
	opts       TUIOptions
	supervisor *Supervisor
	mu         sync.Mutex
	logs       []string
	maxLogs    int
	statusMsg  string
}

// NewTUIDashboard constructs a new TUIDashboard.
func NewTUIDashboard(opts TUIOptions, s *Supervisor) *TUIDashboard {
	if opts.Stdout == nil {
		opts.Stdout = os.Stdout
	}
	if opts.Stdin == nil {
		opts.Stdin = os.Stdin
	}
	if opts.RefreshRate <= 0 {
		opts.RefreshRate = 500 * time.Millisecond
	}
	if opts.ProjectName == "" {
		opts.ProjectName = "Loy Project"
	}

	return &TUIDashboard{
		opts:       opts,
		supervisor: s,
		maxLogs:    8,
	}
}

// Write implements io.Writer to pipe subprocess logs into the rolling activity buffer.
func (d *TUIDashboard) Write(p []byte) (n int, err error) {
	lines := strings.Split(strings.TrimRight(string(p), "\n"), "\n")
	for _, l := range lines {
		if strings.TrimSpace(l) != "" {
			d.AddLog(l)
		}
	}
	return len(p), nil
}

// AddLog appends a message to the rolling log window.
func (d *TUIDashboard) AddLog(line string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.logs = append(d.logs, line)
	if len(d.logs) > d.maxLogs {
		d.logs = d.logs[len(d.logs)-d.maxLogs:]
	}
}

// SetStatus updates the temporary status banner.
func (d *TUIDashboard) SetStatus(msg string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.statusMsg = msg
}

// RenderFrame writes a single refreshed dashboard frame to w.
func (d *TUIDashboard) RenderFrame(w io.Writer, snapshots []ProcessStatusSnapshot, uptime time.Duration) {
	d.mu.Lock()
	defer d.mu.Unlock()

	// Header
	_, _ = fmt.Fprintln(w, "================================================================================")
	_, _ = fmt.Fprintf(w, "  🏛️  LOY DEV DASHBOARD — %s (Uptime: %s)\n", d.opts.ProjectName, uptime.Truncate(time.Second))
	_, _ = fmt.Fprintln(w, "================================================================================")

	// Supervised Processes Section
	_, _ = fmt.Fprintln(w, "  Processes:")
	if len(snapshots) == 0 {
		_, _ = fmt.Fprintln(w, "    (no supervised processes)")
	} else {
		for _, p := range snapshots {
			state := "RUNNING"
			if !p.Running {
				state = "STOPPED"
			}
			pidStr := "-"
			if p.PID > 0 {
				pidStr = fmt.Sprintf("%d", p.PID)
			}
			_, _ = fmt.Fprintf(w, "    • [%-10s] Status: %-7s | PID: %-6s | Restarts: %d\n",
				p.Name, state, pidStr, p.Restarts)
		}
	}

	// Status / Notification banner
	if d.statusMsg != "" {
		_, _ = fmt.Fprintf(w, "\n  [Notice] %s\n", d.statusMsg)
	}

	// Logs Section
	_, _ = fmt.Fprintln(w, "\n  Recent Activity:")
	if len(d.logs) == 0 {
		_, _ = fmt.Fprintln(w, "    (waiting for output...)")
	} else {
		for _, l := range d.logs {
			_, _ = fmt.Fprintf(w, "    > %s\n", l)
		}
	}

	// Footer Commands
	_, _ = fmt.Fprintln(w, "--------------------------------------------------------------------------------")
	_, _ = fmt.Fprintln(w, "  [r] Manual Reload  |  [c] Architecture Check  |  [q] Quit")
	_, _ = fmt.Fprintln(w, "================================================================================")
}

// Run starts the interactive TUI event loop until ctx is canceled.
func (d *TUIDashboard) Run(ctx context.Context) error {
	ticker := time.NewTicker(d.opts.RefreshRate)
	defer ticker.Stop()

	startTime := time.Now()

	// Keyboard input pump
	inputCh := make(chan string, 1)
	go func() {
		scanner := bufio.NewScanner(d.opts.Stdin)
		for scanner.Scan() {
			text := strings.TrimSpace(scanner.Text())
			select {
			case inputCh <- text:
			case <-ctx.Done():
				return
			}
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return nil

		case cmd := <-inputCh:
			switch strings.ToLower(cmd) {
			case "q", "quit", "exit":
				return nil
			case "r", "reload":
				d.SetStatus("Triggering manual reload of processes...")
				if d.supervisor != nil {
					go func() {
						_ = d.supervisor.RestartAll()
					}()
				}
			case "c", "check":
				d.SetStatus("Running live architecture boundaries check...")
			}

		case <-ticker.C:
			var snapshots []ProcessStatusSnapshot
			if d.supervisor != nil {
				d.supervisor.mu.Lock()
				for _, p := range d.supervisor.processes {
					if p == nil {
						continue
					}
					snapshots = append(snapshots, ProcessStatusSnapshot{
						Name:    p.Task().Name,
						PID:     p.PID(),
						Running: p.IsRunning(),
					})
				}
				d.supervisor.mu.Unlock()
			}

			// Clear terminal screen and render frame
			if !d.opts.NoColor {
				_, _ = fmt.Fprint(d.opts.Stdout, "\033[H\033[2J")
			}
			d.RenderFrame(d.opts.Stdout, snapshots, time.Since(startTime))
		}
	}
}
