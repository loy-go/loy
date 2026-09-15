---
title: "Best Practices: Transport Handlers & Routing"
description: "How to design HTTP handlers, middleware bindings, and OpenAPI annotations in Loy."
---

Generated via: `loy make handler <name>` or `loy make crud <name>`

Transport Handlers translate external HTTP requests into application service calls, handle response serialization, and map errors to appropriate HTTP status codes.

## Golden Rules

### 1. Zero Business Logic in Handlers
- **DO NOT** execute SQL queries, calculate prices, or make business workflow decisions inside HTTP handlers.
- **DO** delegate all business logic to Application Services:

```go title="internal/candidate/transport/http/handler.go"
package http

import (
	"net/http"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"myapp/internal/candidate/service"
)

type Handler struct {
	service *service.Service
}

func NewHandler(svc *service.Service) (*Handler, error) {
	return &Handler{service: svc}, nil
}

// GetByID handles GET /candidates/:id
// @Summary Get Candidate
// @Tags candidates
// @Param id path int true "Candidate ID"
// @Success 200 {object} domain.Candidate
// @Router /api/v1/candidates/{id} [get]
func (h *Handler) GetByID(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "invalid ID format"})
	}

	candidate, err := h.service.GetByID(c.Context(), id)
	if err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"data": candidate})
}
```

### 2. Standardized OpenAPI Documentation
- Annotate handler receiver methods with standard Swag comment annotations (`@Summary`, `@Tags`, `@Success`, `@Router`).
- Run `loy doc` to automatically extract OpenAPI 3.0 / Swagger specifications into `docs/`.

### 3. Inspecting Registered Endpoints
- Use `loy routes` to statically inspect and verify registered route paths, HTTP methods, and source handler line numbers without booting the application server.
