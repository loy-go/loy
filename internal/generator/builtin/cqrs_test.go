package builtin_test

import (
	"context"
	"strings"
	"testing"

	"github.com/loy-go/loy/internal/generator"
	"github.com/loy-go/loy/internal/generator/builtin"
)

func TestCQRSGenerators(t *testing.T) {
	ctx := context.Background()

	t.Run("command generator", func(t *testing.T) {
		gen := builtin.NewCommandGenerator("github.com/example/app")
		input := generator.Input{
			Name: "create_order",
			Args: map[string]string{
				"fields": "customer_id:string total:float",
			},
		}

		arts, err := gen.Generate(ctx, input)
		if err != nil {
			t.Fatalf("command generate failed: %v", err)
		}
		if len(arts) != 1 {
			t.Fatalf("expected 1 artifact, got %d", len(arts))
		}

		content := string(arts[0].Content)
		if !strings.Contains(content, "CreateOrderCommand struct") {
			t.Errorf("expected CreateOrderCommand struct in output, got:\n%s", content)
		}
		if !strings.Contains(content, "CreateOrderHandler struct") {
			t.Errorf("expected CreateOrderHandler struct in output, got:\n%s", content)
		}
		if !strings.Contains(content, "CustomerId string") {
			t.Errorf("expected CustomerId in output, got:\n%s", content)
		}
	})

	t.Run("query generator", func(t *testing.T) {
		gen := builtin.NewQueryGenerator("github.com/example/app")
		input := generator.Input{
			Name: "get_order",
			Args: map[string]string{
				"fields": "order_number:string total:float status:string",
			},
		}

		arts, err := gen.Generate(ctx, input)
		if err != nil {
			t.Fatalf("query generate failed: %v", err)
		}
		if len(arts) != 1 {
			t.Fatalf("expected 1 artifact, got %d", len(arts))
		}

		content := string(arts[0].Content)
		if !strings.Contains(content, "GetOrderQuery struct") {
			t.Errorf("expected GetOrderQuery struct in output, got:\n%s", content)
		}
		if !strings.Contains(content, "GetOrderView struct") {
			t.Errorf("expected GetOrderView struct in output, got:\n%s", content)
		}
		if !strings.Contains(content, "GetOrderReader interface") {
			t.Errorf("expected GetOrderReader interface in output, got:\n%s", content)
		}
	})

	t.Run("idempotency generator", func(t *testing.T) {
		gen := builtin.NewIdempotencyGenerator("github.com/example/app")
		input := generator.Input{
			Name: "idempotency",
		}

		arts, err := gen.Generate(ctx, input)
		if err != nil {
			t.Fatalf("idempotency generate failed: %v", err)
		}
		if len(arts) != 2 {
			t.Fatalf("expected 2 artifacts (migration + middleware), got %d", len(arts))
		}

		mig := string(arts[0].Content)
		if !strings.Contains(mig, "CREATE TABLE IF NOT EXISTS idempotency_keys") {
			t.Errorf("expected idempotency_keys table in migration")
		}

		mw := string(arts[1].Content)
		if !strings.Contains(mw, "func Idempotency(store IdempotencyStore") {
			t.Errorf("expected Idempotency middleware function in output")
		}
	})
}
