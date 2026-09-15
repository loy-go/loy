---
title: "Self-Healing Agent Loops"
description: "How loy check --format agent produces machine-readable diagnostic prompts that enable AI coding agents to self-correct in one shot."
---

When an autonomous AI agent introduces an architectural boundary violation, standard Go compiler or linter errors fail to provide sufficient context. A message like:
```text
import cycle not allowed
```
or
```text
cannot use &pg.Repo as type service.Repository
```
causes agents to hallucinate workarounds—such as using `any`, type assertions, or package-level variables—which further degrades code health.

Loy introduces the **Agent Self-Healing Feedback Loop** via:
```bash
loy check --format agent
```

---

## Machine-Readable Agent Diagnostic Schema

When architectural violations occur, `loy check --format agent` emits a clean JSON array of `AgentDiagnostic` payloads with exact code rationale and actionable remediation prompts:

```json
[
  {
    "code": "LOY-ARCH-002",
    "severity": "error",
    "file": "internal/order/domain/order.go",
    "line": 14,
    "violation": "domain package github.com/example/app/internal/order/domain imports infrastructure package github.com/example/app/internal/order/repository/pg",
    "rationale": "Domain entities must remain pure Go structs independent of persistence drivers.",
    "remediation": {
      "action": "invert_dependency",
      "prompt": "Remove import of \"internal/order/repository/pg\" from internal/order/domain/order.go. Define a repository interface in the domain package: \n\ntype OrderRepository interface {\n    FindByID(ctx context.Context, id int64) (*Order, error)\n}\n\nThen implement this interface in internal/.../repository/."
    }
  }
]
```

If all architectural rules pass, `loy check --format agent` emits an empty JSON array:
```json
[]
```
and exits with code `0`.

---

## Rule Remediation Action Matrix

Loy's rules engine categorizes violations into standardized remediation actions for LLMs:

| Rule ID | Rule Name | Remediation Action | Generated Agent Guidance |
|---|---|---|---|
| `ARCH-001` | Dependency Cycle | `break_cycle` | Break the cycle by extracting common types into a shared package or using interface inversion. |
| `ARCH-002` | Domain -> Infrastructure | `invert_dependency` | Remove infrastructure import from domain. Define a repository interface in domain. |
| `ARCH-003` | Domain -> Transport | `remove_import` | Remove transport import. Map DTOs to pure domain entities in the transport layer. |
| `ARCH-004` | Application -> Transport | `invert_call_direction` | Transport handlers should invoke application services, never vice versa. |
| `ARCH-005` | Application -> Concrete Adapter | `invert_dependency` | Declare consumer interfaces in application. Inject concrete adapters via `internal/app/wiring.go`. |
| `ARCH-006` | Infrastructure -> Transport | `remove_import` | Remove transport dependencies from infrastructure storage adapters. |
| `ARCH-007` | Transport Direct Persistence | `delegate_to_service` | Remove raw SQL from handlers. Route mutations through application use cases. |
| `ARCH-008` | Framework in Domain | `isolate_domain` | Remove `net/http`, `gorm`, or `database/sql` from domain entities. |
| `ARCH-009` | Framework in Application | `decouple_transport` | Remove Web framework types (`fiber.Ctx`) from application service signatures. |
| `ARCH-010` | Layer Direction Matrix | `realign_layer` | Realign imports to strictly follow `Transport -> Application -> Domain <- Infrastructure`. |
| `ARCH-011` | Forbidden Service Locator | `explicit_wiring` | Remove DI containers (`uber/dig`). Wire constructors explicitly in `wiring.go`. |
| `ARCH-012` | Package-Level Mutable State | `encapsulate_state` | Remove global `var`. Pass state via struct fields initialized in constructors. |
| `ARCH-013` | Monorepo App -> App | `extract_shared_package` | Decouple applications. Extract shared logic into `internal/` or `packages/`. |
| `ARCH-014` | Corrupted Comment Regions | `restore_comment_region` | Restore missing `// loy:region:...` / `// loy:endregion` markers or run `loy upgrade`. |
| `ARCH-015` | Impure Platform Utilities | `relocate_impure_platform` | Relocate network, OS, or DB operations in platform packages to infrastructure. |

---

## Autonomous Agent Integration

### Claude Code & Kilo Workflows

Add this verification command to your `.kilo/commands` or `AGENTS.md`:

```markdown
Before submitting your changes, execute:
$ loy check --format agent

If any violations are returned in the JSON payload, execute the exact 'remediation.prompt' instructions to self-heal your code in one shot.
```

### Git Pre-Commit Hook

Install the automated architecture gate hook:
```bash
loy hook install
```

This ensures that no developer or AI agent can commit code that violates architectural invariants.
