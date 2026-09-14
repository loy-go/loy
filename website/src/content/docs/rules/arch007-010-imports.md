---
title: "ARCH-007–010: Import Restrictions"
description: "Rules restricting driver imports, web frameworks in use cases, and layer bypasses."
---

## ARCH-007: Transport Cannot Bypass Application Layer

- **Severity**: `WARN`
- **Rule**: Transport handlers must not bypass application services to execute repository operations directly.
- **Why**: Bypassing services bypasses authorization, validation, domain event dispatching, and transaction coordination.

---

## ARCH-008: No Database Drivers in Domain Layer

- **Severity**: `ERROR`
- **Rule**: Files in `domain/` must never import low-level database drivers (`database/sql`, `github.com/jackc/pgx/v5`, `github.com/lib/pq`).
- **Why**: Domain models should be plain Go structs using standard Go types (`string`, `int`, `time.Time`, `uuid.UUID`).

---

## ARCH-009: No Web Frameworks in Application Layer

- **Severity**: `ERROR`
- **Rule**: Application services in `service/` must never import web framework packages (`github.com/gofiber/fiber/v2`, `github.com/gin-gonic/gin`, `github.com/go-chi/chi/v5`).
- **Why**: Web frameworks belong strictly in the Transport layer (`transport/http/`). If use cases depend on `fiber.Ctx`, they cannot be tested in isolation or invoked from background workers.

---

## ARCH-010: Disallowed Third-Party Dependencies

- **Severity**: `WARN`
- **Rule**: Flags third-party imports that violate repository tier policies or introduce unapproved framework abstractions.
