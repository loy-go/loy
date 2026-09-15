package auth

import (
	"github.com/golang-jwt/jwt/v5"
)

// UserClaims defines authenticated user claims contained in the JWT.
type UserClaims struct {
	UserID int64    `json:"user_id"`
	OrgID  string   `json:"org_id,omitempty"`
	Roles  []string `json:"roles,omitempty"`
	jwt.RegisteredClaims
}
