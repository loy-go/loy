package auth

import (
	"net/http"
	"strings"

	"github.com/gofiber/fiber/v2"
)

const ContextKeyUser = "auth_user"

// RequireAuth returns a middleware requiring a valid Bearer JWT.
func RequireAuth(jwtSvc *JWTService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
				"error": "missing or malformed authorization header",
			})
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		claims, err := jwtSvc.ParseToken(tokenStr)
		if err != nil {
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
				"error": "invalid or expired authentication token",
			})
		}

		c.Locals(ContextKeyUser, claims)
		return c.Next()
	}
}

// RequireRole returns a middleware requiring one of the specified roles.
func RequireRole(roles ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		val := c.Locals(ContextKeyUser)
		claims, ok := val.(*UserClaims)
		if !ok || claims == nil {
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
				"error": "unauthorized",
			})
		}

		for _, required := range roles {
			for _, userRole := range claims.Roles {
				if userRole == required {
					return c.Next()
				}
			}
		}

		return c.Status(http.StatusForbidden).JSON(fiber.Map{
			"error": "forbidden: insufficient permissions",
		})
	}
}
