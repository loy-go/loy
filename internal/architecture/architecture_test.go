package architecture

import (
	"testing"
)

func TestLayerClassifier(t *testing.T) {
	c := NewClassifier("github.com/example/app", map[string]string{
		"github.com/example/app/pkg/custom": "Domain",
	})

	tests := []struct {
		importPath string
		expected   Layer
	}{
		{"github.com/example/app/internal/transport/http", LayerTransport},
		{"github.com/example/app/internal/candidate/api", LayerTransport},
		{"github.com/example/app/internal/handler/user", LayerTransport},
		{"github.com/example/app/views/pages", LayerTransport},
		{"github.com/example/app/internal/views/components", LayerTransport},
		{"github.com/example/app/internal/service/order", LayerApplication},
		{"github.com/example/app/internal/worker/cv", LayerApplication},
		{"github.com/example/app/internal/domain/order", LayerDomain},
		{"github.com/example/app/internal/repository/postgres", LayerInfrastructure},
		{"github.com/example/app/internal/platform/cache", LayerPlatform},
		{"github.com/example/app/pkg/mailer", LayerPlatform},
		{"github.com/example/app/pkg/custom", LayerDomain},
		{"github.com/gin-gonic/gin", LayerUnknown},
		{"github.com/redis/go-redis/v9", LayerUnknown},
		{"google.golang.org/grpc", LayerUnknown},
	}

	for _, tt := range tests {
		got := c.Classify(tt.importPath)
		if got != tt.expected {
			t.Errorf("Classify(%q) = %v, want %v", tt.importPath, got, tt.expected)
		}
	}
}

func TestSuppressionFiltering(t *testing.T) {
	v := Violation{
		RuleID: "ARCH-002",
		File:   "domain/user.go",
		Line:   10,
	}

	t.Run("suppressed with valid reason", func(t *testing.T) {
		s := Suppression{
			RuleID: "ARCH-002",
			Reason: "legacy transition",
			File:   "domain/user.go",
			Line:   9,
			Valid:  true,
		}
		filtered := FilterViolations([]Violation{v}, []Suppression{s})
		if len(filtered) != 0 {
			t.Fatalf("expected violation to be suppressed, got %d", len(filtered))
		}
	})

	t.Run("suppression missing reason produces error", func(t *testing.T) {
		s := Suppression{
			RuleID: "ARCH-002",
			Reason: "",
			File:   "domain/user.go",
			Line:   9,
			Valid:  false,
		}
		filtered := FilterViolations([]Violation{v}, []Suppression{s})
		if len(filtered) != 2 { // original violation + suppression error
			t.Fatalf("expected 2 violations, got %d", len(filtered))
		}
	})

	t.Run("non-suppressible rule produces error", func(t *testing.T) {
		vCycle := Violation{
			RuleID: "ARCH-001",
			File:   "a.go",
			Line:   1,
		}
		s := Suppression{
			RuleID: "ARCH-001",
			Reason: "trying to ignore cycle",
			File:   "a.go",
			Line:   1,
			Valid:  true,
		}
		filtered := FilterViolations([]Violation{vCycle}, []Suppression{s})
		if len(filtered) != 2 { // original cycle violation + non-suppressible error
			t.Fatalf("expected 2 violations, got %d", len(filtered))
		}
	})
}
