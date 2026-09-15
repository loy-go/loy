---
title: "Best Practices: Authorization Policies & Security"
description: "How to implement declarative resource permissions, JWT authentication, and RBAC in Loy."
---

Generated via: `loy make policy <name>` and `loy make auth`

Security in Loy follows two distinct layers:
1. **Authentication & Identity:** Verifying who the caller is (`internal/platform/auth/`).
2. **Authorization Policies:** Determining whether the authenticated caller has permission to perform a specific action on a specific resource (`internal/<context>/policy/`).

## Golden Rules

### 1. Declarative Policies over Inline If-Checks
- **DO NOT** scatter complex role and ownership checks across HTTP handlers.
- **DO** encapsulate access rules inside Policy structs:

```go title="internal/candidate/policy/candidate_policy.go"
package policy

import (
	"context"
	"myapp/internal/candidate/domain"
	"myapp/internal/platform/auth"
)

type CandidatePolicy struct{}

func NewCandidatePolicy() *CandidatePolicy {
	return &CandidatePolicy{}
}

// CanView checks if actor can access candidate data.
func (p *CandidatePolicy) CanView(ctx context.Context, claims *auth.UserClaims, candidate *domain.Candidate) bool {
	// Super admins can view all candidates
	if hasRole(claims.Roles, "admin") {
		return true
	}
	// Recruiters can only view candidates within their organization
	return claims.OrgID == candidate.OrgID
}

// CanDelete checks if actor can remove a candidate.
func (p *CandidatePolicy) CanDelete(ctx context.Context, claims *auth.UserClaims, candidate *domain.Candidate) bool {
	return hasRole(claims.Roles, "admin")
}

func hasRole(roles []string, target string) bool {
	for _, r := range roles {
		if r == target {
			return true
		}
	}
	return false
}
```

### 2. Standard Cryptographic Hashing
- When handling user credentials, use `auth.HashPassword` (backed by bcrypt with default cost 10) or Argon2id. Never store or compare plain text passwords.
- Always perform constant-time hash comparisons to prevent timing attacks.
