---
title: "The Architectural Layer Matrix"
description: "Visual breakdown of layer directions, permitted imports, and platform purity rules."
---

Loy standardizes Clean Architecture for Go backends. Package responsibilities and permissible import directions are strictly governed by **Spec 05** and validated via `loy check` ([ADR-016](/loy/adrs/)).

## 1. Dependency Flow Hierarchy

```
  ┌────────────────────────────────────────────────────────┐
  │                    TRANSPORT LAYER                     │
  │     cmd/api, cmd/worker, transport/http, grpc, ws      │
  └───────────────────────────┬────────────────────────────┘
                              │ imports
                              ▼
  ┌────────────────────────────────────────────────────────┐
  │                   APPLICATION LAYER                    │
  │                service, usecase, job                   │
  └───────────────────────────┬────────────────────────────┘
                              │ imports
                              ▼
  ┌────────────────────────────────────────────────────────┐
  │                      DOMAIN LAYER                      │
  │                  model, entity, event                  │
  └───────────────────────────▲────────────────────────────┘
                              │ implements interfaces
  ┌───────────────────────────┴────────────────────────────┐
  │                  INFRASTRUCTURE LAYER                  │
  │       repository, database, platform/cache, asynq      │
  └────────────────────────────────────────────────────────┘
```

---

## 2. Permitted Import Matrix

| From Layer | May Import Into | Forbidden Imports | Governing Rule |
|---|---|---|---|
| **Transport** | Application, Domain, Transport, Platform | Infrastructure direct invocation | `ARCH-007` |
| **Application** | Domain, Application, Platform | Transport, Concrete Infrastructure | `ARCH-004`, `ARCH-005` |
| **Domain** | Domain, Pure Platform | Infrastructure, Transport, DB Drivers | `ARCH-002`, `ARCH-003`, `ARCH-008` |
| **Infrastructure** | Application, Domain, Infrastructure, Platform | Transport | `ARCH-006` |
| **Platform (`pkg/*`)** | Platform, Standard Library | Domain, Application (no upward imports) | `ARCH-015` |

---

## 3. Platform Purity Governance (`ARCH-015`)

Shared packages located in `pkg/*` or `internal/platform/*` are classified as `LayerPlatform` ([ADR-023](/loy/adrs/)):

- **Pure Platform Packages:** Helpers containing only in-memory calculations (math, strings, crypto hashing, time formatters) may be imported by any layer, including Domain.
- **Impure Platform Packages:** Components importing IO, network, or database drivers (`net/http`, `database/sql`, `os`, `syscall`) must **never** be imported into Domain or Application layers.
