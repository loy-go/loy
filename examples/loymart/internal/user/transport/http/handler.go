package http

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"loymart/internal/user/domain"
	"loymart/internal/user/service"
)

// Handler exposes HTTP transport endpoints for users.
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

// RegisterRoutes attaches user routes to the provided Fiber router.
func (h *Handler) RegisterRoutes(router fiber.Router) {
	group := router.Group("/users")
	group.Get("/", h.List)
	group.Get("/:id", h.GetByID)
	group.Post("/", h.Create)
	group.Put("/:id", h.Update)
	group.Delete("/:id", h.Delete)
}

// List handles GET /users.
// @Summary List users
// @Description Retrieve a paginated list of users
// @Tags users
// @Produce json
// @Param limit query int false "Pagination limit" default(20)
// @Param offset query int false "Pagination offset" default(0)
// @Success 200 {object} map[string]any
// @Router /api/v1/users [get]
func (h *Handler) List(c *fiber.Ctx) error {
	limit := c.QueryInt("limit", 20)
	offset := c.QueryInt("offset", 0)

	items, err := h.service.List(c.Context(), limit, offset)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"data": items})
}

// GetByID handles GET /users/:id.
// @Summary Get User by ID
// @Description Retrieve a single User record by identifier
// @Tags users
// @Produce json
// @Param id path int true "User ID"
// @Success 200 {object} domain.User
// @Failure 404 {object} map[string]string
// @Router /api/v1/users/{id} [get]
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

// Create handles POST /users.
// @Summary Create User
// @Description Persist a new User record
// @Tags users
// @Accept json
// @Produce json
// @Param payload body domain.User true "Payload"
// @Success 201 {object} domain.User
// @Router /api/v1/users [post]
func (h *Handler) Create(c *fiber.Ctx) error {
	var entity domain.User
	if err := c.BodyParser(&entity); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "malformed request body"})
	}

	if err := h.service.Create(c.Context(), &entity); err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(http.StatusCreated).JSON(fiber.Map{"data": entity})
}

// Update handles PUT /users/:id.
// @Summary Update User
// @Description Update an existing User record
// @Tags users
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Param payload body domain.User true "Payload"
// @Success 200 {object} domain.User
// @Router /api/v1/users/{id} [put]
func (h *Handler) Update(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "invalid ID format"})
	}

	var entity domain.User
	if err := c.BodyParser(&entity); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "malformed request body"})
	}
	entity.ID = id

	if err := h.service.Update(c.Context(), &entity); err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"data": entity})
}

// Delete handles DELETE /users/:id.
// @Summary Delete User
// @Description Remove a User record by identifier
// @Tags users
// @Param id path int true "User ID"
// @Success 204
// @Router /api/v1/users/{id} [delete]
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
