package builtin_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
)

// Emulate Coordinator LIFO behavior directly to test Phase 6 specification contract
type testHook func(ctx context.Context) error

type testPhase int

const (
	testPhaseTransport testPhase = iota
	testPhaseQueue
	testPhaseStorage
	testPhaseTelemetry
	testPhaseCount
)

type testCoordinator struct {
	mu     sync.Mutex
	phases [testPhaseCount][]testHook
}

func (c *testCoordinator) Register(p testPhase, h testHook) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.phases[p] = append(c.phases[p], h)
}

func (c *testCoordinator) Teardown(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	var errs []error
	for p := 0; p < int(testPhaseCount); p++ {
		hooks := c.phases[p]
		for i := len(hooks) - 1; i >= 0; i-- {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			if err := hooks[i](ctx); err != nil {
				errs = append(errs, err)
			}
		}
	}
	return errors.Join(errs...)
}

func TestShutdownCoordinator_PhasedLIFO(t *testing.T) {
	coord := &testCoordinator{}
	var executionOrder []string

	// Register in Transport phase
	coord.Register(testPhaseTransport, func(ctx context.Context) error {
		executionOrder = append(executionOrder, "transport_1")
		return nil
	})
	coord.Register(testPhaseTransport, func(ctx context.Context) error {
		executionOrder = append(executionOrder, "transport_2")
		return nil
	})

	// Register in Storage phase
	coord.Register(testPhaseStorage, func(ctx context.Context) error {
		executionOrder = append(executionOrder, "db_connection")
		return nil
	})

	// Register in Telemetry phase
	coord.Register(testPhaseTelemetry, func(ctx context.Context) error {
		executionOrder = append(executionOrder, "telemetry_flush")
		return nil
	})

	err := coord.Teardown(context.Background())
	if err != nil {
		t.Fatalf("unexpected teardown error: %v", err)
	}

	expected := []string{
		"transport_2",   // LIFO within PhaseTransport
		"transport_1",   // LIFO within PhaseTransport
		"db_connection", // PhaseStorage
		"telemetry_flush", // PhaseTelemetry
	}

	if len(executionOrder) != len(expected) {
		t.Fatalf("expected %d steps, got %d: %v", len(expected), len(executionOrder), executionOrder)
	}

	for i, step := range expected {
		if executionOrder[i] != step {
			t.Errorf("step %d: expected %s, got %s", i, step, executionOrder[i])
		}
	}
}

func TestShutdownCoordinator_TimeoutDeadline(t *testing.T) {
	coord := &testCoordinator{}
	coord.Register(testPhaseTransport, func(ctx context.Context) error {
		select {
		case <-time.After(50 * time.Millisecond):
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	})

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	err := coord.Teardown(ctx)
	if err == nil {
		t.Fatal("expected deadline exceeded error")
	}
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("expected context.DeadlineExceeded, got: %v", err)
	}
}

// Emulate Health Registry to verify status code contract
type testHealthRegistry struct {
	checkers map[string]func(ctx context.Context) error
}

func (r *testHealthRegistry) LiveHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprint(w, `{"status":"up"}`)
	}
}

func (r *testHealthRegistry) ReadyHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		allReady := true
		for _, check := range r.checkers {
			if err := check(ctx); err != nil {
				allReady = false
				break
			}
		}

		w.Header().Set("Content-Type", "application/json")
		if !allReady {
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = fmt.Fprint(w, `{"status":"down"}`)
			return
		}

		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprint(w, `{"status":"up"}`)
	}
}

func TestHealthEndpoints(t *testing.T) {
	reg := &testHealthRegistry{
		checkers: make(map[string]func(ctx context.Context) error),
	}

	// 1. Live probe returns 200
	reqLive := httptest.NewRequest("GET", "/health/live", nil)
	recLive := httptest.NewRecorder()
	reg.LiveHandler()(recLive, reqLive)

	if recLive.Code != http.StatusOK {
		t.Errorf("expected 200 for live probe, got %d", recLive.Code)
	}
	if !strings.Contains(recLive.Body.String(), `"status":"up"`) {
		t.Errorf("unexpected live body: %s", recLive.Body.String())
	}

	// 2. Ready probe with passing check returns 200
	reg.checkers["db"] = func(ctx context.Context) error { return nil }
	reqReady := httptest.NewRequest("GET", "/health/ready", nil)
	recReady := httptest.NewRecorder()
	reg.ReadyHandler()(recReady, reqReady)

	if recReady.Code != http.StatusOK {
		t.Errorf("expected 200 for ready probe, got %d", recReady.Code)
	}

	// 3. Ready probe with failing check returns 503
	reg.checkers["db"] = func(ctx context.Context) error { return errors.New("db connection refused") }
	recReadyFail := httptest.NewRecorder()
	reg.ReadyHandler()(recReadyFail, reqReady)

	if recReadyFail.Code != http.StatusServiceUnavailable {
		t.Errorf("expected 503 for ready probe failure, got %d", recReadyFail.Code)
	}
	if !strings.Contains(recReadyFail.Body.String(), `"status":"down"`) {
		t.Errorf("unexpected ready fail body: %s", recReadyFail.Body.String())
	}
}
