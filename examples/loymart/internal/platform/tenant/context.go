package tenant

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"regexp"

	"github.com/gofiber/fiber/v2"
)

type tenantKeyType struct{}

var tenantKey = tenantKeyType{}

var tenantIDRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]{1,64}$`)

// ErrMissingTenant is returned when no tenant identifier is present.
var ErrMissingTenant = errors.New("missing tenant identifier")

// ErrInvalidTenant is returned when tenant identifier format is invalid.
var ErrInvalidTenant = errors.New("invalid tenant identifier format")

// WithTenantID returns a context with tenant ID attached.
func WithTenantID(ctx context.Context, tenantID string) context.Context {
	return context.WithValue(ctx, tenantKey, tenantID)
}

// GetTenantID retrieves the current tenant ID from context.
func GetTenantID(ctx context.Context) (string, bool) {
	val, ok := ctx.Value(tenantKey).(string)
	return val, ok && val != ""
}

// SetLocalTenantRLS sets the transaction-local PostgreSQL RLS session variable.
func SetLocalTenantRLS(ctx context.Context, tx *sql.Tx, tenantID string) error {
	if !tenantIDRegex.MatchString(tenantID) {
		return ErrInvalidTenant
	}
	_, err := tx.ExecContext(ctx, "SELECT set_config('app.current_tenant_id', $1, true)", tenantID)
	if err != nil {
		return fmt.Errorf("setting rls tenant session: %w", err)
	}
	return nil
}

// FiberMiddleware extracts tenant ID from header or query and injects into context.
func FiberMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		tenantID := c.Get("X-Tenant-ID")
		if tenantID == "" {
			tenantID = c.Query("tenant_id")
		}
		if tenantID != "" {
			if !tenantIDRegex.MatchString(tenantID) {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error": "invalid tenant id format",
				})
			}
			ctx := WithTenantID(c.UserContext(), tenantID)
			c.SetUserContext(ctx)
		}
		return c.Next()
	}
}
