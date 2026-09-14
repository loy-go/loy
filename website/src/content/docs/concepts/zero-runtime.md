---
title: "Zero Runtime Invariant"
description: "How Loy delivers high developer velocity without introducing runtime framework lock-in."
---

Most web frameworks (Ruby on Rails, Laravel, Django, Spring) act as heavy runtime dependencies. When you write applications with them, your code is tightly coupled to the framework's internal classes, base controllers, and ORM abstractions.

Loy takes the opposite approach: **Build-Time First, Zero Runtime Lock-In** ([ADR-002](/adrs/)).

## Ordinary Go Output

When you run `loy new` or `loy make crud`:
- Generated files are standard, plain Go source files (`.go`).
- Generated packages import only standard library packages and mature, un-wrapped open-source libraries (`github.com/jackc/pgx/v5`, `github.com/gofiber/fiber/v2`, `github.com/hibiken/asynq`).
- **Your application never imports `github.com/loy-go/loy`**.

---

## What Happens If You Delete the Loy Binary?

If your team decides to stop using the Loy CLI:
1. Your project continues to compile with standard `go build ./...`.
2. Your tests continue to run with standard `go test ./...`.
3. Your deployment pipelines run unchanged.
4. You own 100% of the generated code without any framework dependencies.

Loy serves as your **accelerator**, not your jailer.
