---
title: "Self-Healing Agent Loops & Automated Fixes"
description: "How loy check --format agent and loy check --fix enable autonomous AI coding agents to diagnose and heal architectural violations."
---

When an autonomous AI agent introduces an architectural boundary violation, standard Go compiler or linter errors fail to provide actionable context:
```text
cannot use &pg.Repo as type service.Repository
```
Such vague errors often lead LLMs into a "hallucination loop"—introducing reflection, type casts, or package-level variables, which further degrades code health.

Loy introduces the **Self-Healing Agent Loop** through two complementary mechanisms:
1. **`loy check --format agent`**: Emits machine-readable `AgentDiagnostic` JSON payloads with step-by-step remediation instructions.
2. **`loy check --fix`**: Automatically rewrites Go ASTs to remediate deterministic architectural violations (such as `ARCH-005` interface extraction) without manual coding.

---

## 1. The Closed-Loop Self-Healing Architecture

```text
┌────────────────────────────────────────────────────────────────────────┐
│ 1. CODE GENERATION / MODIFICATION                                      │
│    Agent creates or edits code (manually or via loy make / MCP).       │
└───────────────────────────────────┬────────────────────────────────────┘
                                    │
                                    ▼
┌────────────────────────────────────────────────────────────────────────┐
│ 2. ARCHITECTURAL BOUNDARY CHECK                                        │
│    Agent executes: loy check --format agent                            │
└───────────────────────────────────┬────────────────────────────────────┘
                                    │
                    ┌───────────────┴───────────────┐
                    │ Violations Detected?          │
                    └───────┬───────────────┬───────┘
                       No   │               │ Yes
                            ▼               ▼
                   Exit Code 0      ┌────────────────────────────────────┐
                 [Commit Changes]   │ 3. ATTEMPT DETERMINISTIC AUTO-FIX  │
                                    │    Execute: loy check --fix        │
                                    └───────────────┬────────────────────┘
                                                    │
                                    ┌───────────────┴───────────────┐
                                    │ Resolved?                     │
                                    └───────┬───────────────┬───────┘
                                       Yes  │               │ No
                                            ▼               ▼
                                   Exit Code 0      ┌────────────────────┐
                                 [Commit Changes]   │ 4. LLM REMEDIATION │
                                                    │ Parse remediation  │
                                                    │ prompt from JSON   │
                                                    └─────────┬──────────┘
                                                              │
                                                              ▼
                                                    [Apply DST Mutation]
                                                              │
                                                              ▼
                                                    [Verify: loy check]
```

---

## 2. Automated Remediation (`loy check --fix`)

For standard architectural anti-patterns—such as an Application service directly importing a concrete Infrastructure adapter (`ARCH-005`)—Loy can automatically remediate the violation.

### 2.1 Before `loy check --fix` (Violation)
An AI agent generated a service that directly imports the PostgreSQL repository:

```go title="internal/order/service/service.go"
package service

import (
	"context"
	// ARCH-005 VIOLATION: Application layer imports concrete Infrastructure package!
	"github.com/example/bookstore/internal/order/repository"
)

type Service struct {
	repo *repository.PostgresRepository
}

func NewService(repo *repository.PostgresRepository) *Service {
	return &Service{repo: repo}
}
```

Running `loy check` fails:
```text
ERROR [ARCH-005] internal/order/service/service.go:5
  application package directly imports infrastructure package github.com/example/bookstore/internal/order/repository
```

### 2.2 Running `loy check --fix`
Execute:

```bash
loy check --fix
```

Output:
```text
Auto-remediated 1 architectural violation(s). All architecture rules passed.
```

### 2.3 After `loy check --fix` (Clean Architecture Seam)
Loy rewrites the file's Concrete Syntax Tree (`dst`):
1. Removes the prohibited `repository` infrastructure import.
2. Rewrites struct fields and constructor parameters to depend on an interface.
3. Automatically declares the consumer-owned `Repository` interface directly inside the application layer:

```go title="internal/order/service/service.go"
package service

import (
	"context"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// Repository is the consumer interface defined by the application layer.
type Repository interface {
}
```

---

## 3. Machine-Readable Diagnostic Schema (`--format agent`)

When complex violations require conceptual reasoning (e.g. breaking a cyclic dependency or decoupling a domain aggregate), `loy check --format agent` emits a clean JSON array:

```json
[
  {
    "code": "LOY-ARCH-002",
    "severity": "error",
    "file": "internal/order/domain/order.go",
    "line": 14,
    "violation": "domain package imports infrastructure package repository",
    "rationale": "Domain entities must remain pure Go structs independent of persistence drivers.",
    "remediation": {
      "action": "invert_dependency",
      "prompt": "Remove import of \"internal/order/repository\" from internal/order/domain/order.go. Define a repository interface in domain.go: \n\ntype OrderRepository interface {\n    FindByID(ctx context.Context, id int64) (*Order, error)\n}\n\nThen implement this interface in internal/order/repository/pg_adapter.go."
    }
  }
]
```

When no violations exist, it emits an empty JSON array `[]` and exits with code `0`.

---

## 4. Rule Remediation Action Matrix

| Rule ID | Rule Name | Remediation Action | Generated Agent Guidance |
|---|---|---|---|
| **`ARCH-001`** | Dependency Cycle | `break_cycle` | Break the cycle by extracting common types into a shared package or using interface inversion. |
| **`ARCH-002`** | Domain -> Infrastructure | `invert_dependency` | Remove infrastructure import from domain. Define a repository interface in domain. |
| **`ARCH-003`** | Domain -> Transport | `remove_import` | Remove transport import. Map DTOs to pure domain entities in the transport layer. |
| **`ARCH-004`** | Application -> Transport | `invert_call_direction` | Transport handlers should invoke application services, never vice versa. |
| **`ARCH-005`** | Application -> Concrete Adapter | `invert_dependency` | Run `loy check --fix` or declare consumer interfaces in application. |
| **`ARCH-006`** | Infrastructure -> Transport | `remove_import` | Remove transport dependencies from infrastructure storage adapters. |
| **`ARCH-007`** | Transport Direct Persistence | `delegate_to_service` | Remove raw SQL from handlers. Route mutations through application use cases. |
| **`ARCH-008`** | Framework in Domain | `isolate_domain` | Remove `net/http`, `gorm`, or `database/sql` from domain entities. |
| **`ARCH-009`** | Framework in Application | `decouple_transport` | Remove Web framework types (`fiber.Ctx`) from application service signatures. |
| **`ARCH-010`** | Layer Direction Matrix | `realign_layer` | Realign imports to strictly follow configured architectural DAG. |
| **`ARCH-011`** | Forbidden Service Locator | `explicit_wiring` | Remove DI containers (`uber/dig`). Wire constructors explicitly in `wiring.go`. |
| **`ARCH-012`** | Package-Level Mutable State | `encapsulate_state` | Remove global `var`. Pass state via struct fields initialized in constructors. |
| **`ARCH-013`** | Monorepo App -> App | `extract_shared_package` | Decouple applications. Extract shared logic into `internal/` or `packages/`. |
| **`ARCH-014`** | Corrupted Comment Regions | `restore_comment_region` | Restore missing `// loy:region:...` / `// loy:endregion` markers. |
| **`ARCH-015`** | Impure Platform Utilities | `relocate_impure_platform` | Relocate network, OS, or DB operations in platform packages to infrastructure. |
