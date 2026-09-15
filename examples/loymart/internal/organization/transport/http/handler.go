package http

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"loymart/internal/organization/domain"
	"loymart/internal/organization/service"
)

// Handler exposes HTTP transport endpoints for organizations.
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

// RegisterRoutes attaches organization routes to the provided Fiber router.
func (h *Handler) RegisterRoutes(router fiber.Router) {
	group := router.Group("/organizations")
	group.Get("/", h.List)
	group.Get("/:id", h.GetByID)
	group.Post("/", h.Create)
	group.Put("/:id", h.Update)
	group.Delete("/:id", h.Delete)
}

// List handles GET /organizations.
// @Summary List organizations
// @Description Retrieve a paginated list of organizations
// @Tags organizations
// @Produce json
// @Param limit query int false "Pagination limit" default(20)
// @Param offset query int false "Pagination offset" default(0)
// @Success 200 {object} map[string]any
// @Router /api/v1/organizations [get]
func (h *Handler) List(c *fiber.Ctx) error {
	limit := c.QueryInt("limit", 20)
	offset := c.QueryInt("offset", 0)

	items, err := h.service.List(c.Context(), limit, offset)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"data": items})
}

// GetByID handles GET /organizations/:id.
// @Summary Get Organization by ID
// @Description Retrieve a single Organization record by identifier
// @Tags organizations
// @Produce json
// @Param id path int true "Organization ID"
// @Success 200 {object} domain.Organization
// @Failure 404 {object} map[string]string
// @Router /api/v1/organizations/{id} [get]
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

// Create handles POST /organizations.
// @Summary Create Organization
// @Description Persist a new Organization record
// @Tags organizations
// @Accept json
// @Produce json
// @Param payload body domain.Organization true "Payload"
// @Success 201 {object} domain.Organization
// @Router /api/v1/organizations [post]
func (h *Handler) Create(c *fiber.Ctx) error {
	var entity domain.Organization
	if err := c.BodyParser(&entity); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "malformed request body"})
	}

	if err := h.service.Create(c.Context(), &entity); err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(http.StatusCreated).JSON(fiber.Map{"data": entity})
}

// Update handles PUT /organizations/:id.
// @Summary Update Organization
// @Description Update an existing Organization record
// @Tags organizations
// @Accept json
// @Produce json
// @Param id path int true "Organization ID"
// @Param payload body domain.Organization true "Payload"
// @Success 200 {object} domain.Organization
// @Router /api/v1/organizations/{id} [put]
func (h *Handler) Update(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "invalid ID format"})
	}

	var entity domain.Organization
	if err := c.BodyParser(&entity); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "malformed request body"})
	}
	entity.ID = id

	if err := h.service.Update(c.Context(), &entity); err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"data": entity})
}

// Delete handles DELETE /organizations/:id.
// @Summary Delete Organization
// @Description Remove a Organization record by identifier
// @Tags organizations
// @Param id path int true "Organization ID"
// @Success 204
// @Router /api/v1/organizations/{id} [delete]
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
