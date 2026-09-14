---
title: "Architecture Enforcement Overview"
description: "Static architecture validation, boundary checks, and CI enforcement with loy check."
---

Architecture erosion occurs when developers under deadline pressure take shortcuts: importing infrastructure drivers into domain logic, introducing circular package dependencies, or scattering database queries across transport handlers.

Loy's **Architecture Engine** (`loy check`) runs static analysis against your project's AST and package dependency graph ([ADR-016](/loy/adrs/)), catching violations in milliseconds.

## Running the Architecture Validator

```bash
loy check
```

### Validator Flags

- **`--deep`**: Perform deep semantic type analysis via `golang.org/x/tools/go/packages` in addition to fast AST inspection.
- **`--strict`**: Treat architectural warnings as fatal errors (recommended for CI build gates).
- **`--json`**: Emit violations in machine-readable diagnostic schema.

```bash
# Recommended CI pipeline validation
loy check --strict --json
```

---

## The 15 Standard Rules Catalog

| Rule ID | Category | Severity | Description |
|---|---|---|---|
| **ARCH-001** | Cycles | ERROR | Circular package dependencies are forbidden. |
| **ARCH-002** | Layer Direction | ERROR | Domain layer must never import Infrastructure layer. |
| **ARCH-003** | Layer Direction | ERROR | Domain layer must never import Transport layer. |
| **ARCH-004** | Layer Direction | ERROR | Application layer must never import Transport layer. |
| **ARCH-005** | Layer Direction | WARN | Application layer should not import concrete Infrastructure adapters. |
| **ARCH-006** | Layer Direction | ERROR | Infrastructure layer must never import Transport layer. |
| **ARCH-007** | Layer Direction | WARN | Transport handlers must not bypass service layer to invoke repositories directly. |
| **ARCH-008** | Imports | ERROR | Domain layer must never import direct database driver packages (`database/sql`, `pgx`). |
| **ARCH-009** | Imports | ERROR | Application service layer must never import web frameworks (`fiber`, `gin`, `chi`). |
| **ARCH-010** | Imports | WARN | Disallowed third-party package dependencies outside approved tier. |
| **ARCH-011** | Governance | ERROR | Forbidden reflection service locators and global singletons. |
| **ARCH-012** | Governance | ERROR | Package-level mutable state (`var db *sql.DB`) is prohibited. |
| **ARCH-013** | Governance | ERROR | Application-to-application cross-boundary imports in monorepos are forbidden. |
| **ARCH-014** | Governance | WARN | Unmanaged code in generator-managed regions. |
| **ARCH-015** | Governance | ERROR | Domain and Application layers must not import impure platform packages (IO/network/database). |
