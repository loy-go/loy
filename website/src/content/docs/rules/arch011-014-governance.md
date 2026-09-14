---
title: "ARCH-011–014: Governance & State"
description: "Rules preventing global mutable state, reflection service locators, and cross-application leakage."
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
