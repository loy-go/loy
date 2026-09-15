---
title: "ARCH-001: Package Dependency Cycles Prohibited"
description: "Why Go forbids circular dependencies, how loy check detects import cycles, and step-by-step strategies to break them."
---

- **Rule ID**: `ARCH-001`
- **Code**: `LOY-ARCH-001`
- **Severity**: `ERROR` (Non-suppressible)
- **Category**: Dependency Graph DAG Integrity

The Go compiler famously refuses to compile any project containing circular package imports (`import cycle not allowed`). More importantly, dependency cycles indicate architectural entanglement: two or more packages cannot be tested, reasoned about, or deployed independently.

Loy mandates that package dependencies form a strict **Directed Acyclic Graph (DAG)**.

---

## What a Cycle Looks Like

A dependency cycle occurs when Package A depends on Package B, and Package B directly or indirectly depends back on Package A:

```text
┌─────────────────┐       imports       ┌─────────────────┐
│  package user   │ ──────────────────> │  package order  │
└─────────────────┘                     └─────────────────┘
         ▲                                       │
         │                imports                │
         └───────────────────────────────────────┘
```

### Real-World Bug Scenario

Consider an e-commerce application:
1. `package user` needs to check if a user has active orders:
   ```go title="internal/user/domain/user.go"
   package user

   import "myapp/internal/order/domain" // User imports Order

   type User struct {
       ID     int64
       Orders []order.Order
   }
   ```
2. Meanwhile, `package order` needs to know the user's name:
   ```go title="internal/order/domain/order.go"
   package order

   import "myapp/internal/user/domain" // Order imports User! -> CYCLE!

   type Order struct {
       ID       int64
       Customer user.User
   }
   ```

When you try to compile this code, Go immediately aborts:
```text
import cycle not allowed
package myapp/internal/user/domain
    imports myapp/internal/order/domain
    imports myapp/internal/user/domain
```

---

## How `loy check` Catches Cycles

Loy analyzes your package graph at the AST level before compilation and pinpoints the exact chain of offending imports:

```bash
loy check
```

Output:
```text
ERROR [LOY-ARCH-001] internal/user/domain/user.go:3
  cyclic dependency detected: internal/user/domain -> internal/order/domain -> internal/user/domain
  Detail: package internal/user/domain forms an illegal dependency cycle
  Hint: break the cycle by extracting common types or using dependency inversion
```

With `--format agent`, Loy emits the self-healing prompt payload:
```json
{
  "code": "LOY-ARCH-001",
  "severity": "error",
  "file": "internal/user/domain/user.go",
  "line": 3,
  "violation": "cyclic dependency detected: internal/user/domain -> internal/order/domain -> internal/user/domain",
  "rationale": "Package dependency cycles prevent compilation and violate the Directed Acyclic Graph (DAG) package architecture.",
  "remediation": {
    "action": "break_cycle",
    "prompt": "Break the cyclic dependency involving internal/user/domain. Extract shared types and interfaces into a lower-level leaf package or invert dependencies using interfaces."
  }
}
```

---

## Step-by-Step Remediation Strategies

There are two primary ways to eliminate dependency cycles cleanly:

### Strategy 1: Dependency Inversion via Interfaces (Recommended)

Instead of having `order` import `user.User` directly, declare a consumer-owned interface in `order` or pass primitive IDs:

```go title="internal/order/domain/order.go (Refactored)"
package order

// Decoupled: store UserID as an identifier, not the concrete User struct!
type Order struct {
    ID     int64
    UserID int64
    Total  float64
}

// If user details are needed, define a narrow interface where consumed:
type UserInfoProvider interface {
    GetCustomerName(ctx context.Context, userID int64) (string, error)
}
```

Now, `order` no longer imports `user`! The cycle is completely broken.

---

### Strategy 2: Extract a Shared Leaf Package

If two packages share a common type (e.g. an address model or shared money/currency type):

```text
Before (Cycle):
   user <────────────> order

After (DAG):
   user ──────┐   ┌────── order
              ▼   ▼
        shared/money
```

1. Create a leaf package: `internal/common/types` (or `internal/platform/currency`).
2. Move the shared type into the leaf package.
3. Both `user` and `order` import the leaf package. Because the leaf package imports neither, the graph is a strict DAG.

---

## Why ARCH-001 Is Non-Suppressible

Some rules in Loy allow temporary suppression via comments (`// loy:suppress:ARCH-005: legacy migration`).

However, **`ARCH-001` cannot be suppressed under any circumstances**. The Go compiler itself physically forbids circular imports; suppressing the rule in Loy would still result in a broken, non-compilable Go binary.
