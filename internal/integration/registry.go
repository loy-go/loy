package integration

import (
	"fmt"
	"sync"
)

// Adapter defines the contract for an integration adapter.
type Adapter interface {
	Name() string
	Capability() Capability
	Tier() int // 1, 2, or 3
	Description() string
}

// Registry provides process-scoped lookup and registration of adapters.
// Avoids global mutable registration per ADR-005 and AGENTS.md.
type Registry struct {
	mu       sync.RWMutex
	adapters map[Capability]map[string]Adapter
	defaults map[Capability]string
}

// NewRegistry initializes an explicit, process-scoped integration registry.
func NewRegistry() *Registry {
	return &Registry{
		adapters: make(map[Capability]map[string]Adapter),
		defaults: make(map[Capability]string),
	}
}

// Register registers an adapter under its capability.
func (r *Registry) Register(adapter Adapter) error {
	if adapter == nil {
		return fmt.Errorf("cannot register nil adapter")
	}
	cap := adapter.Capability()
	if !cap.IsValid() {
		return fmt.Errorf("invalid capability: %s", cap)
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.adapters[cap]; !exists {
		r.adapters[cap] = make(map[string]Adapter)
	}

	name := adapter.Name()
	r.adapters[cap][name] = adapter

	// If no default set for this capability, use the first registered
	if _, hasDefault := r.defaults[cap]; !hasDefault {
		r.defaults[cap] = name
	}

	return nil
}

// SetDefault sets the default adapter name for a capability.
func (r *Registry) SetDefault(cap Capability, name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if bucket, exists := r.adapters[cap]; exists {
		if _, ok := bucket[name]; ok {
			r.defaults[cap] = name
			return nil
		}
	}
	return fmt.Errorf("adapter %q not found for capability %s", name, cap)
}

// Get resolves an adapter by capability and optional name. If name is empty, returns the default.
func (r *Registry) Get(cap Capability, name string) (Adapter, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	bucket, exists := r.adapters[cap]
	if !exists || len(bucket) == 0 {
		return nil, fmt.Errorf("no adapters registered for capability %s", cap)
	}

	if name == "" {
		name = r.defaults[cap]
	}

	adapter, ok := bucket[name]
	if !ok {
		return nil, fmt.Errorf("adapter %q not found for capability %s", name, cap)
	}

	return adapter, nil
}

// List returns all registered adapters for a capability.
func (r *Registry) List(cap Capability) []Adapter {
	r.mu.RLock()
	defer r.mu.RUnlock()

	bucket, exists := r.adapters[cap]
	if !exists {
		return nil
	}

	result := make([]Adapter, 0, len(bucket))
	for _, a := range bucket {
		result = append(result, a)
	}
	return result
}

// BaseAdapter is a reusable struct satisfying Adapter.
type BaseAdapter struct {
	AdapterName        string
	AdapterCapability  Capability
	AdapterTier        int
	AdapterDescription string
}

func (b BaseAdapter) Name() string        { return b.AdapterName }
func (b BaseAdapter) Capability() Capability { return b.AdapterCapability }
func (b BaseAdapter) Tier() int           { return b.AdapterTier }
func (b BaseAdapter) Description() string { return b.AdapterDescription }

// DefaultRegistry creates a pre-populated registry with Tier 1 integrations.
func DefaultRegistry() *Registry {
	r := NewRegistry()

	_ = r.Register(BaseAdapter{
		AdapterName:        "fiber",
		AdapterCapability:  CapHTTP,
		AdapterTier:        1,
		AdapterDescription: "High-performance Express-inspired HTTP framework adapter",
	})
	_ = r.Register(BaseAdapter{
		AdapterName:        "postgres",
		AdapterCapability:  CapDatabase,
		AdapterTier:        1,
		AdapterDescription: "PostgreSQL pgxpool adapter with embedded goose and sqlc",
	})
	_ = r.Register(BaseAdapter{
		AdapterName:        "valkey",
		AdapterCapability:  CapCache,
		AdapterTier:        1,
		AdapterDescription: "Valkey / Redis client adapter for caching and transient state",
	})
	_ = r.Register(BaseAdapter{
		AdapterName:        "asynq",
		AdapterCapability:  CapQueue,
		AdapterTier:        1,
		AdapterDescription: "Asynq distributed asynchronous background task and worker engine",
	})
	_ = r.Register(BaseAdapter{
		AdapterName:        "otel",
		AdapterCapability:  CapTelemetry,
		AdapterTier:        1,
		AdapterDescription: "OpenTelemetry SDK tracer and metrics provider",
	})

	return r
}
