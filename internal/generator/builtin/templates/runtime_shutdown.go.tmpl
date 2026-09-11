package shutdown

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

// Phase defines the sequential bucket in which teardown hooks are executed.
type Phase int

const (
	PhaseTransport Phase = iota // HTTP listeners, gRPC servers
	PhaseQueue                  // Background workers, job queues
	PhaseStorage                // Databases, caches, connection pools
	PhaseTelemetry              // Telemetry flush, loggers
	PhaseCount                  // Total phases
)

// Hook is a cleanup function called during shutdown.
type Hook func(ctx context.Context) error

// Coordinator manages ordered, phased, LIFO cleanup execution.
type Coordinator struct {
	mu     sync.Mutex
	phases [PhaseCount][]Hook
}

// NewCoordinator creates a new shutdown Coordinator.
func NewCoordinator() *Coordinator {
	return &Coordinator{}
}

// Register adds a teardown hook to a designated phase.
func (c *Coordinator) Register(phase Phase, hook Hook) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if phase < 0 || phase >= PhaseCount {
		phase = PhaseStorage
	}
	c.phases[phase] = append(c.phases[phase], hook)
}

// Teardown executes registered hooks phase-by-phase (PhaseTransport -> PhaseTelemetry),
// executing hooks in reverse registration order (LIFO) within each phase.
func (c *Coordinator) Teardown(ctx context.Context) error {
	c.mu.Lock()
	var phasesCopy [PhaseCount][]Hook
	for p := 0; p < int(PhaseCount); p++ {
		phasesCopy[p] = make([]Hook, len(c.phases[p]))
		copy(phasesCopy[p], c.phases[p])
	}
	c.mu.Unlock()

	var errs []error
	for p := 0; p < int(PhaseCount); p++ {
		hooks := phasesCopy[p]
		for i := len(hooks) - 1; i >= 0; i-- {
			if ctx.Err() != nil {
				return errors.Join(append(errs, fmt.Errorf("shutdown deadline exceeded: %w", ctx.Err()))...)
			}
			if err := hooks[i](ctx); err != nil {
				errs = append(errs, err)
			}
		}
	}

	return errors.Join(errs...)
}
