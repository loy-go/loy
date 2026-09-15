---
title: "Migrating from Raw Gin to Loy Architecture"
description: "Refactor untyped monolithic Gin routes and global database handles into clean architecture vertical slices."
---

Gin is one of the most popular HTTP web frameworks in the Go ecosystem. However, as Gin applications grow, common architectural pitfalls emerge:
- **Monolithic Route Handlers**: Route registration, validation, business rules, and SQL queries mixed inside single `gin.HandlerFunc` closures.
- **Global Mutable State**: Storing `*gorm.DB` or `*sql.DB` in package globals or `gin.Context` keys (`c.Set("db", db)`).
- **Leaky Transport**: Database models imported directly into HTTP handler JSON tags.

Loy supports Gin as a first-class transport adapter while enforcing clean architectural separation via static AST analysis (`loy check`).

---

## 1. Before: The Monolithic Gin Pattern

In typical raw Gin projects, handlers directly query databases and mutate global state:

```go
// anti-pattern: handlers doing everything
func CreateUserHandler(c *gin.Context) {
    var req struct {
        Email string `json:"email" binding:"required"`
    }
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }

    // Direct database access inside HTTP handler (violates ARCH-007)
    var user User
    if err := globalDB.Where("email = ?", req.Email).First(&user).Error; err == nil {
        c.JSON(409, gin.H{"error": "email taken"})
        return
    }

    newUser := User{Email: req.Email}
    globalDB.Create(&newUser)
    c.JSON(201, newUser)
}
```

---

## 2. After: Clean Architecture with Loy

Loy structures applications into four strict layers: `Transport -> Application -> Domain <- Infrastructure`.

### 1. Domain Entity (`internal/user/domain/user.go`)
```go
package domain

import "time"

type User struct {
    ID        int64     `json:"id"`
    Email     string    `json:"email"`
    CreatedAt time.Time `json:"created_at"`
}

type UserRepository interface {
    GetByEmail(ctx context.Context, email string) (*User, error)
    Create(ctx context.Context, user *User) error
}
```

### 2. Application Service (`internal/user/service/service.go`)
```go
package service

import (
    "context"
    "errors"
    "my-app/internal/user/domain"
)

type Service struct {
    repo domain.UserRepository
}

func NewService(repo domain.UserRepository) (*Service, error) {
    if repo == nil {
        return nil, errors.New("repository dependency is required")
    }
    return &Service{repo: repo}, nil
}

func (s *Service) Register(ctx context.Context, email string) (*domain.User, error) {
    existing, _ := s.repo.GetByEmail(ctx, email)
    if existing != nil {
        return nil, errors.New("email already taken")
    }
    user := &domain.User{Email: email}
    if err := s.repo.Create(ctx, user); err != nil {
        return nil, err
    }
    return user, nil
}
```

### 3. Gin Transport Handler (`internal/user/transport/http/handler.go`)
```go
package http

import (
    "net/http"
    "github.com/gin-gonic/gin"
    "my-app/internal/user/service"
)

type Handler struct {
    svc *service.Service
}

func NewHandler(svc *service.Service) (*Handler, error) {
    return &Handler{svc: svc}, nil
}

func (h *Handler) RegisterRoutes(router *gin.RouterGroup) {
    router.POST("/users", h.Create)
}

func (h *Handler) Create(c *gin.Context) {
    var req struct {
        Email string `json:"email" binding:"required"`
    }
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    user, err := h.svc.Register(c.Request.Context(), req.Email)
    if err != nil {
        c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusCreated, gin.H{"data": user})
}
```

---

## 3. Verifying the Migration with `loy check`

Run Loy's static architecture analyzer to verify that:
- `internal/user/domain` imports zero infrastructure or web frameworks (`ARCH-002`, `ARCH-008`).
- `internal/user/service` does not import Gin (`ARCH-009`).
- Handlers do not bypass services to execute database queries (`ARCH-007`).

```bash
loy check
```

Output:
```text
All architecture rules passed. (0 violations across 4 layers)
```
