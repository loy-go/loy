package benchmarks_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/go-chi/chi/v5"
	"github.com/gofiber/fiber/v2"
)

// setupFiber initializes Fiber router with a typical Loy route layout.
func setupFiber() *fiber.App {
	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
	})
	api := app.Group("/api/v1")
	api.Get("/users", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok", "items": 10})
	})
	api.Get("/users/:id", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"id": c.Params("id"), "name": "Alice"})
	})
	return app
}

// setupChi initializes Chi router with an identical route layout.
func setupChi() *chi.Mux {
	r := chi.NewRouter()
	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/users", func(w http.ResponseWriter, req *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"status":"ok","items":10}`))
		})
		r.Get("/users/{id}", func(w http.ResponseWriter, req *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"id":"` + chi.URLParam(req, "id") + `","name":"Alice"}`))
		})
	})
	return r
}

// setupGin initializes Gin router with an identical route layout.
func setupGin() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	g := gin.New()
	api := g.Group("/api/v1")
	api.GET("/users", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "items": 10})
	})
	api.GET("/users/:id", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"id": c.Param("id"), "name": "Alice"})
	})
	return g
}

// setupNetHTTP initializes standard Go 1.22+ ServeMux with an identical route layout.
func setupNetHTTP() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/users", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok","items":10}`))
	})
	mux.HandleFunc("GET /api/v1/users/{id}", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"id":"` + req.PathValue("id") + `","name":"Alice"}`))
	})
	return mux
}

// TestBenchmarkSmoke ensures test harnesses initialize cleanly.
func TestBenchmarkSmoke(t *testing.T) {
	f := setupFiber()
	if f == nil {
		t.Fatal("fiber setup failed")
	}
	c := setupChi()
	if c == nil {
		t.Fatal("chi setup failed")
	}
	g := setupGin()
	if g == nil {
		t.Fatal("gin setup failed")
	}
	n := setupNetHTTP()
	if n == nil {
		t.Fatal("nethttp setup failed")
	}
}

// BenchmarkFiber tests Fiber request handling and JSON response dispatch.
func BenchmarkFiber(b *testing.B) {
	app := setupFiber()
	req := httptest.NewRequest("GET", "/api/v1/users/42", nil)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		resp, err := app.Test(req, -1)
		if err != nil || resp.StatusCode != http.StatusOK {
			b.Fatalf("fiber request failed: %v", err)
		}
		_ = resp.Body.Close()
	}
}

// BenchmarkChi tests Chi router request dispatch.
func BenchmarkChi(b *testing.B) {
	router := setupChi()
	req := httptest.NewRequest("GET", "/api/v1/users/42", nil)
	w := httptest.NewRecorder()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		w.Body.Reset()
		router.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			b.Fatalf("chi request failed with status %d", w.Code)
		}
	}
}

// BenchmarkGin tests Gin router request dispatch.
func BenchmarkGin(b *testing.B) {
	router := setupGin()
	req := httptest.NewRequest("GET", "/api/v1/users/42", nil)
	w := httptest.NewRecorder()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		w.Body.Reset()
		router.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			b.Fatalf("gin request failed with status %d", w.Code)
		}
	}
}

// BenchmarkNetHTTP tests standard library ServeMux request dispatch.
func BenchmarkNetHTTP(b *testing.B) {
	mux := setupNetHTTP()
	req := httptest.NewRequest("GET", "/api/v1/users/42", nil)
	w := httptest.NewRecorder()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		w.Body.Reset()
		mux.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			b.Fatalf("net/http request failed with status %d", w.Code)
		}
	}
}

func BenchmarkParallel(b *testing.B) {
	b.Run("fiber", func(b *testing.B) {
		app := setupFiber()
		b.ReportAllocs()
		b.RunParallel(func(pb *testing.PB) {
			req := httptest.NewRequest("GET", "/api/v1/users/42", nil)
			for pb.Next() {
				resp, _ := app.Test(req, -1)
				_, _ = io.Copy(io.Discard, resp.Body)
				_ = resp.Body.Close()
			}
		})
	})

	b.Run("chi", func(b *testing.B) {
		router := setupChi()
		b.ReportAllocs()
		b.RunParallel(func(pb *testing.PB) {
			req := httptest.NewRequest("GET", "/api/v1/users/42", nil)
			w := httptest.NewRecorder()
			for pb.Next() {
				w.Body.Reset()
				router.ServeHTTP(w, req)
			}
		})
	})

	b.Run("gin", func(b *testing.B) {
		router := setupGin()
		b.ReportAllocs()
		b.RunParallel(func(pb *testing.PB) {
			req := httptest.NewRequest("GET", "/api/v1/users/42", nil)
			w := httptest.NewRecorder()
			for pb.Next() {
				w.Body.Reset()
				router.ServeHTTP(w, req)
			}
		})
	})

	b.Run("nethttp", func(b *testing.B) {
		mux := setupNetHTTP()
		b.ReportAllocs()
		b.RunParallel(func(pb *testing.PB) {
			req := httptest.NewRequest("GET", "/api/v1/users/42", nil)
			w := httptest.NewRecorder()
			for pb.Next() {
				w.Body.Reset()
				mux.ServeHTTP(w, req)
			}
		})
	})
}
