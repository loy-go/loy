package integration_test

import (
	"testing"

	"github.com/uloydev/loy/internal/integration"
)

func TestCapabilityValidity(t *testing.T) {
	validCaps := []integration.Capability{
		integration.CapHTTP,
		integration.CapDatabase,
		integration.CapCache,
		integration.CapQueue,
		integration.CapRPC,
		integration.CapTemplate,
		integration.CapTelemetry,
		integration.CapAssets,
	}

	for _, c := range validCaps {
		if !c.IsValid() {
			t.Errorf("expected capability %s to be valid", c)
		}
		if c.String() == "" {
			t.Errorf("expected non-empty string for capability %s", c)
		}
	}

	invalid := integration.Capability("invalid-cap")
	if invalid.IsValid() {
		t.Errorf("expected invalid-cap to be invalid")
	}
}

func TestRegistryOperations(t *testing.T) {
	reg := integration.NewRegistry()

	// Register nil adapter
	if err := reg.Register(nil); err == nil {
		t.Errorf("expected error registering nil adapter")
	}

	// Register invalid capability adapter
	badAdapter := integration.BaseAdapter{
		AdapterName:       "bad",
		AdapterCapability: integration.Capability("unknown"),
	}
	if err := reg.Register(badAdapter); err == nil {
		t.Errorf("expected error registering adapter with unknown capability")
	}

	// Register valid adapter
	fiberAdapter := integration.BaseAdapter{
		AdapterName:        "fiber",
		AdapterCapability:  integration.CapHTTP,
		AdapterTier:        1,
		AdapterDescription: "Fiber HTTP transport",
	}
	if err := reg.Register(fiberAdapter); err != nil {
		t.Fatalf("unexpected error registering adapter: %v", err)
	}

	// Query default
	got, err := reg.Get(integration.CapHTTP, "")
	if err != nil {
		t.Fatalf("unexpected error resolving default HTTP adapter: %v", err)
	}
	if got.Name() != "fiber" {
		t.Errorf("expected adapter fiber, got %s", got.Name())
	}
	if got.Tier() != 1 {
		t.Errorf("expected tier 1, got %d", got.Tier())
	}
	if got.Description() != "Fiber HTTP transport" {
		t.Errorf("unexpected description: %s", got.Description())
	}

	// List
	list := reg.List(integration.CapHTTP)
	if len(list) != 1 {
		t.Errorf("expected 1 adapter in list, got %d", len(list))
	}

	// Non-existent capability
	if _, err := reg.Get(integration.CapDatabase, ""); err == nil {
		t.Errorf("expected error getting unregistered capability")
	}
	if emptyList := reg.List(integration.CapDatabase); len(emptyList) != 0 {
		t.Errorf("expected empty list for unregistered capability")
	}
}

func TestDefaultRegistry(t *testing.T) {
	reg := integration.DefaultRegistry()

	expectedCaps := []struct {
		cap  integration.Capability
		name string
	}{
		{integration.CapHTTP, "fiber"},
		{integration.CapDatabase, "postgres"},
		{integration.CapCache, "valkey"},
		{integration.CapQueue, "asynq"},
		{integration.CapTelemetry, "otel"},
	}

	for _, tc := range expectedCaps {
		adapter, err := reg.Get(tc.cap, tc.name)
		if err != nil {
			t.Errorf("failed to get default adapter for %s: %v", tc.cap, err)
			continue
		}
		if adapter.Name() != tc.name {
			t.Errorf("expected name %s, got %s", tc.name, adapter.Name())
		}
		if adapter.Tier() != 1 {
			t.Errorf("expected Tier 1 for %s, got %d", tc.name, adapter.Tier())
		}
	}
}
