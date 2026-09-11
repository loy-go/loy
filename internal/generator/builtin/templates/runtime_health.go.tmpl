package health

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"
)

// Checker performs an individual subsystem health/readiness check.
type Checker interface {
	Check(ctx context.Context) error
}

// CheckerFunc adapts a function to the Checker interface.
type CheckerFunc func(ctx context.Context) error

func (f CheckerFunc) Check(ctx context.Context) error {
	return f(ctx)
}

// Registry manages registered readiness checkers.
type Registry struct {
	mu       sync.RWMutex
	checkers map[string]Checker
}

// NewRegistry constructs an empty health check Registry.
func NewRegistry() *Registry {
	return &Registry{
		checkers: make(map[string]Checker),
	}
}

// Register adds a named health checker to the registry.
func (r *Registry) Register(name string, checker Checker) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.checkers[name] = checker
}

// Status represents the health check response payload.
type Status struct {
	Status string            `json:"status"`
	Checks map[string]string `json:"checks,omitempty"`
}

// LiveHandler returns 200 OK immediately if the process is alive.
func (r *Registry) LiveHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(Status{Status: "up"})
	}
}

// ReadyHandler evaluates all registered health checkers.
func (r *Registry) ReadyHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		r.mu.RLock()
		checkersCopy := make(map[string]Checker, len(r.checkers))
		for k, v := range r.checkers {
			checkersCopy[k] = v
		}
		r.mu.RUnlock()

		ctx := req.Context()
		results := make(map[string]string, len(checkersCopy))
		isReady := true

		for name, checker := range checkersCopy {
			if err := checker.Check(ctx); err != nil {
				results[name] = err.Error()
				isReady = false
			} else {
				results[name] = "ok"
			}
		}

		w.Header().Set("Content-Type", "application/json")
		if !isReady {
			w.WriteHeader(http.StatusServiceUnavailable)
			_ = json.NewEncoder(w).Encode(Status{
				Status: "down",
				Checks: results,
			})
			return
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(Status{
			Status: "up",
			Checks: results,
		})
	}
}
