---
title: "Best Practices: Architecture Enforcement in CI/CD"
description: "How to integrate loy check and pre-commit hooks into your development and deployment pipelines."
---

Enforced via: `loy check` and `loy hook install` ([ADR-011](/loy/adrs/), [ADR-016](/loy/adrs/))

The **Architecture Engine** (`loy check`) prevents architecture erosion by continuously validating layer directions, import boundaries, and platform purity.

## Golden Rules

### 1. Shift Left: Local Pre-Commit Hooks
- Run `loy hook install` on local workstations.
- Every commit automatically runs `loy check --quiet` in under 50 milliseconds, blocking invalid imports before code reaches the git remote.

### 2. Strict CI Build Gates
- In GitHub Actions / GitLab CI, run `loy check --strict` as the first step in your PR verification workflow:

```yaml title=".github/workflows/ci.yml"
steps:
  - uses: actions/checkout@v4
  - uses: actions/setup-go@v5
    with:
      go-version: '1.24'

  # Architecture verification gate
  - name: Enforce Architecture Rules
    run: |
      go install github.com/loy-go/loy/cmd/loy@latest
      loy check --strict
```

### 3. Documented Rule Suppressions
- If a temporary architecture exception is required during a legacy migration, use structured comment suppressions with explicit reasons:
  ```go
  // loy:suppress:ARCH-002 reason="legacy migration phase 2"
  ```
- Suppressions missing a valid `reason` or applied to non-suppressible rules (`ARCH-010`) are rejected.
