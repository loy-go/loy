---
title: "Best Practices: Architecture Enforcement in CI/CD"
description: "How to integrate loy check, git lifecycle hooks, drift diffing, and GitHub Actions annotations into continuous delivery pipelines."
---

Enforced via:
```bash
# Validate architecture
loy check [--deep] [--strict] [--format text|json|github|agent]

# Install local Git pre-commit lifecycle hook
loy hook install
```

Architecture rules written in wiki pages or confluence documents fail because they rely on manual human vigilance. Under deadline pressure, engineers accidentally import persistence drivers into domain models or web routers into application use cases.

Loy replaces manual code review gatekeeping with **automated, compile-time architectural enforcement** ([ADR-016](/loy/adrs/)).

---

## The Two-Phase Enforcement Engine

Loy's architecture engine operates in two complementary phases:

```text
┌────────────────────────────────────────────────────────────────────────┐
│ PHASE 1: AST IMPORT & TOPOGRAPHY PARSING (< 50ms)                      │
│ - Parses Go syntax trees without invoking the compiler.                │
│ - Validates 4-layer import directions (ARCH-001 to ARCH-010).          │
│ - Validates comment region integrity (ARCH-014).                       │
│ - Runs on every git commit in microseconds.                            │
└───────────────────────────────────┬────────────────────────────────────┘
                                    │ Optional --deep flag
                                    ▼
┌────────────────────────────────────────────────────────────────────────┐
│ PHASE 2: DEEP TYPE ANALYSIS VIA GO/PACKAGES (CI Gate)                  │
│ - Evaluates concrete interface implementations across packages.        │
│ - Detects subtle service locator anti-patterns (ARCH-011).             │
│ - Enforces platform purity across transitive type assignments (ARCH-15)│
└────────────────────────────────────────────────────────────────────────┘
```

---

## 1. Shift Left: Local Pre-Commit Git Hooks

Run `loy hook install` once when setting up a repository:

```bash
loy hook install
```

This installs a lightweight `.git/hooks/pre-commit` hook that executes Phase 1 AST verification on every `git commit`. If a developer or AI assistant introduces a layer violation, the commit is rejected **before code ever reaches the Git remote**:

```text
git commit -m "add order logic"
ERROR [LOY-ARCH-002] internal/order/domain/order.go:14
  domain package internal/order/domain imports infrastructure package internal/order/repository/pg
  Hint: define repository interface in domain or application layer and implement in infrastructure
Commit aborted by Loy Architecture Gate.
```

---

## 2. GitHub Actions Workflow Annotations (`--format github`)

When running inside GitHub Actions, pass `--format github` to render interactive inline annotations directly on the Pull Request diff:

```yaml title=".github/workflows/ci.yml"
name: Continuous Integration

on:
  pull_request:
    branches: [main]
  push:
    branches: [main]

jobs:
  architecture:
    name: Architecture Enforcement
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.24'

      - name: Install Loy
        run: go install github.com/loy-go/loy/cmd/loy@latest

      - name: Verify Architecture Boundaries
        run: loy check --strict --format github
```

Violations appear directly in GitHub's code review UI, with exact file line pointers, rule IDs, and remediation suggestions.

---

## 3. Pull Request Architecture Drift Diffing

Detect structural package erosion between branches:

```yaml
      - name: Check Architecture Drift
        run: |
          loy graph --diff origin/main --format markdown > drift.md
          cat drift.md
```

If new packages or violations were introduced, Loy outputs a Markdown report suitable for posting as a PR review comment.

---

## 4. Structured Rule Suppressions

During large legacy migrations, teams occasionally need temporary exceptions for specific lines. Loy supports **explicit, documented comment suppressions**:

```go
// loy:suppress:ARCH-002 reason="migrating legacy orders; tracked in issue #412"
import "myapp/internal/legacy/database"
```

### Strict Suppression Rules:
1. **Mandatory Reason**: Suppressions missing a valid `reason="..."` attribute are rejected with `LOY-ARCH-099`.
2. **Non-Suppressible Invariants**: Critical rules cannot be suppressed under any circumstances:
   - `ARCH-001` (Dependency Cycles)
   - `ARCH-010` (Global Layer Matrix Direction)
   - `ARCH-013` (Monorepo App-to-App dependencies)

---

## Standard CLI Exit Codes

`loy check` adheres strictly to standard UNIX exit codes:

| Exit Code | Meaning | CI Impact |
|---|---|---|
| `0` | **Success**: Zero architectural violations found. | Pipeline succeeds. |
| `1` | **Validation Failure**: One or more architectural rules failed. | Pipeline halts with failure. |
| `2` | **Usage Error**: Unknown flag or invalid arguments. | Pipeline configuration error. |
| `3` | **System Error**: Internal panic or I/O failure. | Infrastructure fault. |
