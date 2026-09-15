---
title: "Best Practices: Request DTOs & Resource Transformers"
description: "How to validate input payloads, prevent mass assignment, and transform API responses in Loy."
---

Generated via: `loy make request <name>` and `loy make resource <name>`

DTOs (Data Transfer Objects) prevent internal database columns from leaking into public APIs and protect applications from mass assignment security vulnerabilities.

## Golden Rules

### 1. Decouple Request DTOs from Domain Entities
- **DO NOT** bind untrusted incoming JSON payloads directly onto domain entities:
  ```go
  // ANTI-PATTERN: Mass assignment vulnerability!
  var entity domain.User
  _ = c.BodyParser(&entity) // Attacker can overwrite entity.IsAdmin or entity.OrgID
  ```
- **DO** bind to a dedicated `CreateUserRequest` struct and validate fields:

```go title="internal/user/transport/http/request/create_user_request.go"
package request

import "errors"

type CreateUserRequest struct {
	Email string `json:"email"`
	Name  string `json:"name"`
}

func (r *CreateUserRequest) Validate() error {
	if r.Email == "" {
		return errors.New("email is required")
	}
	if r.Name == "" {
		return errors.New("name is required")
	}
	return nil
}
```

### 2. Transform Responses with Resource DTOs
- Format dates, omit sensitive fields (password hashes, internal flags), and provide computed properties via Resource transformers:

```go title="internal/user/transport/http/resource/user_resource.go"
package resource

import "myapp/internal/user/domain"

type UserResource struct {
	ID        int64  `json:"id"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	CreatedAt string `json:"created_at"`
}

func TransformUser(u *domain.User) UserResource {
	return UserResource{
		ID:        u.ID,
		Email:     u.Email,
		Name:      u.Name,
		CreatedAt: u.CreatedAt.UTC().Format("2006-01-02T15:04:05Z07:00"),
	}
}
```
