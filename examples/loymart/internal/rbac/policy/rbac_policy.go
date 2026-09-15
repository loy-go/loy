package policy

import (
	"context"

	"loymart/internal/user/domain"
)

type contextKey string

// RoleContextKey stores the actor role in context.
const RoleContextKey contextKey = "role"

// RbacPolicy defines authorization checks for user and organization resources.
type RbacPolicy struct{}

// NewRbacPolicy constructs a new authorization policy.
func NewRbacPolicy() *RbacPolicy {
	return &RbacPolicy{}
}

func getActorRole(ctx context.Context) string {
	if r, ok := ctx.Value(RoleContextKey).(string); ok {
		return r
	}
	return ""
}

// CanView determines if the caller has view permission.
func (p *RbacPolicy) CanView(ctx context.Context, actorID int64, entity *domain.User) bool {
	if entity == nil {
		return false
	}
	return entity.ID == actorID || getActorRole(ctx) == "admin"
}

// CanEdit determines if the caller has edit permission.
func (p *RbacPolicy) CanEdit(ctx context.Context, actorID int64, entity *domain.User) bool {
	if entity == nil {
		return false
	}
	return entity.ID == actorID || getActorRole(ctx) == "admin"
}

// CanDelete determines if the caller has delete permission.
func (p *RbacPolicy) CanDelete(ctx context.Context, actorID int64, entity *domain.User) bool {
	if entity == nil {
		return false
	}
	return getActorRole(ctx) == "admin"
}
