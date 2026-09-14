---
title: "ARCH-011–015: Governance, State & Platform Purity"
description: "Rules preventing global mutable state, reflection service locators, cross-application leakage, and impure platform imports."
---

## ARCH-011: Forbidden Service Locators

- **Severity**: `ERROR`
- **Rule**: Prohibits reflection-based service locators and runtime dynamic container lookups (`container.Get("userService")`).
- **Remediation**: Use explicit constructor injection in `internal/app/wiring.go`.

---

## ARCH-012: Package-Level Mutable State

- **Severity**: `ERROR`
- **Rule**: Package-level mutable variables (`var db *sql.DB`, `var currentUser *User`, `var config Config`) are strictly prohibited.
- **Why**: Global state causes concurrency race conditions, breaks parallel test isolation (`t.Parallel()`), and creates hidden coupling between packages.
- **Remediation**: Encapsulate state within struct instances and pass them via constructor injection or `context.Context`.

---

## ARCH-013: Cross-Application Monorepo Leakage

- **Severity**: `ERROR`
- **Rule**: In multi-app workspaces (`apps/api` and `apps/worker`), one application must never import code directly from another application (`apps/api/internal/...` imported by `apps/worker`).
- **Remediation**: Move shared domain models or utility packages into `packages/shared/`.

---

## ARCH-014: Unmanaged Code in Generated Regions

- **Severity**: `WARN`
- **Rule**: Detects handwritten code placed inside generator-managed comment regions (`// loy:region:...`) that could be overwritten during subsequent scaffolding runs.

---

## ARCH-015: Platform Utility Purity

- **Severity**: `ERROR`
- **Rule**: Domain and Application layers must never import impure platform or `pkg/*` packages that perform direct IO, network requests, or database driver operations (`net/http`, `database/sql`, `os`, `syscall`).
- **Why**: Domain logic must remain clean, deterministic, and free of side-effects.
- **Remediation**: Keep pure helpers (string manipulation, math, hashing) in platform/pkg, and move impure client implementations (mailers, database loggers, HTTP clients) to the Infrastructure layer behind domain-defined interfaces.
