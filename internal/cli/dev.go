package cli

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"github.com/loy-go/loy/internal/dev"
	"github.com/loy-go/loy/internal/filesystem"
	"github.com/loy-go/loy/internal/process"
)

func newDevCmd(fs filesystem.FileSystem, runner process.Runner) *cobra.Command {
	var (
		debounce    time.Duration
		gracePeriod time.Duration
		apiOnly     bool
		tui         bool
	)

	cmd := &cobra.Command{
		Use:   "dev [path]",
		Short: "Start multi-process live development server with hot reload",
		Long:  `loy dev supervises application processes (API server, background workers) and reloads on file changes using native fsnotify watching.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			rootDir := "."
			if len(args) > 0 {
				rootDir = args[0]
			}

			cliOpts := GetOptions(cmd.Context())

			tasks := dev.DiscoverTasks(rootDir, fs)
			if apiOnly {
				var filtered []dev.ProcessTask
				for _, t := range tasks {
					if t.Name == "api" || t.Name == "web" {
						filtered = append(filtered, t)
					}
				}
				if len(filtered) == 0 {
					return fmt.Errorf("no API or Web task found in %s (cmd/api or cmd/web directory not found)", rootDir)
				}
				tasks = filtered
			} else if len(tasks) == 0 {
				// Default fallback
				tasks = append(tasks, dev.ProcessTask{
					Name:    "app",
					Command: "go",
					Args:    []string{"run", "."},
					Dir:     rootDir,
				})
			}

			sOpts := dev.SupervisorOptions{
				RootDir:     rootDir,
				Tasks:       tasks,
				Debounce:    debounce,
				GracePeriod: gracePeriod,
				Stdout:      cmd.OutOrStdout(),
				NoColor:     cliOpts.NoColor,
			}

			ctx, stop := signal.NotifyContext(cmd.Context(), os.Interrupt, syscall.SIGTERM)
			defer stop()

			if tui {
				tuiOpts := dev.TUIOptions{
					ProjectName: rootDir,
					Stdout:      cmd.OutOrStdout(),
					Stdin:       cmd.InOrStdin(),
					RefreshRate: 500 * time.Millisecond,
					NoColor:     cliOpts.NoColor,
				}
				dashboard := dev.NewTUIDashboard(tuiOpts, nil)
				sOpts.Stdout = dashboard

				supervisor, err := dev.NewSupervisor(sOpts, fs)
				if err != nil {
					return fmt.Errorf("initializing dev supervisor: %w", err)
				}
				dashboard.SetSupervisor(supervisor)

				childCtx, cancel := context.WithCancel(ctx)
				defer cancel()

				supervisorErr := make(chan error, 1)
				go func() {
					supervisorErr <- supervisor.Run(childCtx)
					cancel()
				}()

				_ = dashboard.Run(childCtx)
				cancel()
				if err := <-supervisorErr; err != nil && !errors.Is(err, context.Canceled) {
					return err
				}
				return nil
			}

			supervisor, err := dev.NewSupervisor(sOpts, fs)
			if err != nil {
				return fmt.Errorf("initializing dev supervisor: %w", err)
			}

			if err := supervisor.Run(ctx); err != nil && !errors.Is(err, context.Canceled) {
				return err
			}
			return nil
		},
	}

	cmd.Flags().DurationVar(&debounce, "debounce", 200*time.Millisecond, "File change debounce window")
	cmd.Flags().DurationVar(&gracePeriod, "grace-period", 3*time.Second, "Grace period before sending SIGKILL")
	cmd.Flags().BoolVar(&apiOnly, "api-only", false, "Run only the API server process")
	cmd.Flags().BoolVar(&tui, "tui", false, "Start interactive full-terminal development dashboard")

	return cmd
}
