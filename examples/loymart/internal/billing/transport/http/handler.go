package http

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"loymart/internal/billing/domain"
	"loymart/internal/billing/service"
)

// Handler exposes HTTP transport endpoints for billings.
type Handler struct {
	service *service.Service
}

// NewHandler constructs a new Handler.
func NewHandler(svc *service.Service) (*Handler, error) {
	if svc == nil {
		return nil, fmt.Errorf("service dependency is required")
	}
	return &Handler{service: svc}, nil
}

// RegisterRoutes attaches billing routes to the provided Fiber router.
func (h *Handler) RegisterRoutes(router fiber.Router) {
	group := router.Group("/billings")
	group.Get("/", h.List)
	group.Get("/:id", h.GetByID)
	group.Post("/", h.Create)
	group.Put("/:id", h.Update)
	group.Delete("/:id", h.Delete)
}

// List handles GET /billings.
// @Summary List billings
// @Description Retrieve a paginated list of billings
// @Tags billings
// @Produce json
// @Param limit query int false "Pagination limit" default(20)
// @Param offset query int false "Pagination offset" default(0)
// @Success 200 {object} map[string]any
// @Router /api/v1/billings [get]
func (h *Handler) List(c *fiber.Ctx) error {
	limit := c.QueryInt("limit", 20)
	offset := c.QueryInt("offset", 0)

	items, err := h.service.List(c.Context(), limit, offset)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"data": items})
}

// GetByID handles GET /billings/:id.
// @Summary Get Billing by ID
// @Description Retrieve a single Billing record by identifier
// @Tags billings
// @Produce json
// @Param id path int true "Billing ID"
// @Success 200 {object} domain.Billing
// @Failure 404 {object} map[string]string
// @Router /api/v1/billings/{id} [get]
func (h *Handler) GetByID(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "invalid ID format"})
	}

	item, err := h.service.GetByID(c.Context(), id)
	if err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"data": item})
}

// Create handles POST /billings.
// @Summary Create Billing
// @Description Persist a new Billing record
// @Tags billings
// @Accept json
// @Produce json
// @Param payload body domain.Billing true "Payload"
// @Success 201 {object} domain.Billing
// @Router /api/v1/billings [post]
func (h *Handler) Create(c *fiber.Ctx) error {
	var entity domain.Billing
	if err := c.BodyParser(&entity); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "malformed request body"})
	}

	if err := h.service.Create(c.Context(), &entity); err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(http.StatusCreated).JSON(fiber.Map{"data": entity})
}

// Update handles PUT /billings/:id.
// @Summary Update Billing
// @Description Update an existing Billing record
// @Tags billings
// @Accept json
// @Produce json
// @Param id path int true "Billing ID"
// @Param payload body domain.Billing true "Payload"
// @Success 200 {object} domain.Billing
// @Router /api/v1/billings/{id} [put]
func (h *Handler) Update(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "invalid ID format"})
	}

	var entity domain.Billing
	if err := c.BodyParser(&entity); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "malformed request body"})
	}
	entity.ID = id

	if err := h.service.Update(c.Context(), &entity); err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"data": entity})
}

// Delete handles DELETE /billings/:id.
// @Summary Delete Billing
// @Description Remove a Billing record by identifier
// @Tags billings
// @Param id path int true "Billing ID"
// @Success 204
// @Router /api/v1/billings/{id} [delete]
func (h *Handler) Delete(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "invalid ID format"})
	}

	if err := h.service.Delete(c.Context(), id); err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.SendStatus(http.StatusNoContent)
}
