package model_test

import (
	"testing"

	"github.com/loy-go/loy/internal/generator/model"
)

func TestOwnership_String(t *testing.T) {
	if model.DeveloperOwned.String() != "developer" {
		t.Errorf("expected developer, got %s", model.DeveloperOwned.String())
	}
	if model.GeneratedOwned.String() != "generated" {
		t.Errorf("expected generated, got %s", model.GeneratedOwned.String())
	}
	if model.MixedOwned.String() != "mixed" {
		t.Errorf("expected mixed, got %s", model.MixedOwned.String())
	}
}
