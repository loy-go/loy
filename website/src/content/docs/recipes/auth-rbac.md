---
title: "Recipe: JWT Authentication & Role-Based Access Control"
description: "How to implement secure password hashing, JWT token rotation, and role authorization."
---

Loy provides an opt-in, zero-external-framework authentication kit using standard Go libraries (`golang.org/x/crypto/bcrypt` and `github.com/golang-jwt/jwt/v5`) ([ADR-022](/loy/adrs/)).

## 1. Scaffolding the Auth Foundation

Run `loy make auth`:

```bash
loy make auth
```

Scaffolded artifacts in `internal/platform/auth/`:
- **`password.go`**: Safe password hashing and constant-time comparison via bcrypt.
- **`jwt.go`**: HMAC-SHA256 signed token generation and validation.
- **`claims.go`**: Standard `UserClaims` (`UserID`, `OrgID`, `Roles`).
- **`middleware.go`**: `RequireAuth()`, `RequireRole()`, and `TenantContext()`.

## 2. Token Issuance on Login

```go
func (h *Handler) Login(c *fiber.Ctx) error {
    var req LoginRequest
    if err := c.BodyParser(&req); err != nil {
        return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "invalid payload"})
    }

    user, err := h.userService.Authenticate(c.Context(), req.Email, req.Password)
    if err != nil {
        return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "invalid credentials"})
    }

    token, err := h.jwtService.MintToken(user.ID, user.OrgID, user.Roles)
    if err != nil {
        return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "failed to sign token"})
    }

    return c.JSON(fiber.Map{"token": token})
}
```

## 3. Protecting Route Groups with RBAC

Attach authorization middlewares to route groups:

```go
api := router.Group("/api/v1")

// All routes require valid Bearer token
api.Use(auth.RequireAuth(jwtSvc))

// Only users with "admin" or "recruiter" role can create jobs
api.Post("/jobs", auth.RequireRole("admin", "recruiter"), jobHandler.Create)

// Extract authenticated user claims in handlers
func (h *Handler) GetProfile(c *fiber.Ctx) error {
    claims := c.Locals("auth_user").(*auth.UserClaims)
    return c.JSON(fiber.Map{"user_id": claims.UserID, "org_id": claims.OrgID})
}
```
