package route_test

import (
	"testing"

	"github.com/loy-go/loy/internal/filesystem"
	"github.com/loy-go/loy/internal/route"
)

func TestInspectRoutes(t *testing.T) {
	memFS := filesystem.NewMemFileSystem()
	_ = memFS.MkdirAll("/proj/internal/transport/http", 0755)

	src := `package http

import "net/http"

type Handler struct{}

func Register(mux *http.ServeMux, r Router, h *Handler) {
	mux.HandleFunc("GET /api/v1/orders", h.ListOrders)
	mux.HandleFunc("POST /api/v1/orders", h.CreateOrder)
	r.Get("/health", h.Health)
	r.Delete("/api/v1/orders/:id", h.DeleteOrder)
}
`
	_ = memFS.WriteFile("/proj/internal/transport/http/routes.go", []byte(src), 0644)

	routes, err := route.InspectRoutes(memFS, "/proj")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(routes) != 4 {
		t.Fatalf("expected 4 routes, got %d: %+v", len(routes), routes)
	}

	foundMap := make(map[string]string)
	for _, r := range routes {
		foundMap[r.Method+" "+r.Path] = r.Handler
	}

	expected := map[string]string{
		"GET /api/v1/orders":         "h.ListOrders",
		"POST /api/v1/orders":        "h.CreateOrder",
		"GET /health":                "h.Health",
		"DELETE /api/v1/orders/:id":  "h.DeleteOrder",
	}

	for k, v := range expected {
		if foundMap[k] != v {
			t.Errorf("expected %s -> %s, got: %s", k, v, foundMap[k])
		}
	}
}
