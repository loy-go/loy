# ADR-022: Baseline Authentication & Session Kit

## Status
Accepted

## Context
Every production application requires user authentication, password hashing, JWT token generation, and role-based access control. Relying on copy-pasted boilerplate creates security vulnerabilities and inconsistent implementations.

## Decision
1. Provide `loy make auth` to generate a zero-dependency baseline authentication foundation.
2. Scaffolds `internal/platform/auth/` containing:
   - Password hashing and constant-time verification using `golang.org/x/crypto/bcrypt`.
   - Signed JWT token minting and validation using `github.com/golang-jwt/jwt/v5`.
   - Typed `UserClaims` (`user_id`, `org_id`, `roles`).
   - Transport authentication and RBAC middlewares (`RequireAuth`, `RequireRole`).
3. Maintain ADR-002: generated code is completely developer-owned and uses standard Go libraries.

## Consequences
- Secure, standardized authentication primitives out-of-the-box.
- Zero vendor lock-in or proprietary runtime auth frameworks.
- Immediate readiness for SaaS multi-tenancy and RBAC policies.

---

[Back to ADR Index](./README.md) | [Back to Documentation Index](../00-INDEX.md)
