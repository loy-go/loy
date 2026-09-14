---
title: "Prerequisites & Doctor"
description: "System prerequisites, toolchain dependencies, and using loy doctor."
---

Loy is designed to minimize external dependencies. Most operations require only standard Go and Git.

## Required Toolchain

- **Go Compiler**: Go 1.24.0 or higher.
- **Git**: Version 2.0+ for workspace resolution and repository management.

---

## Optional Developer Tooling

Depending on which integrations you use in your project, you may install these mature open-source tools:

| Tool | Purpose | Installation |
|---|---|---|
| **sqlc** | Type-safe Go code generation from SQL | `go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest` |
| **Docker** | Containerization and local database orchestration | [docker.com](https://www.docker.com/) |
| **golangci-lint** | High-performance Go linter | `go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest` |
| **Templ** | HTML component templating (for fullstack projects) | `go install github.com/a-h/templ/cmd/templ@latest` |

---

## Verifying Environment with `loy doctor`

The `loy doctor` command inspects your system and project environment, diagnosing issues and providing remediation hints:

```bash
loy doctor
```

### Doctor Diagnostic Flags

- **`--strict`**: Treat all warnings as fatal errors (useful in CI pipelines).
- **`--json`**: Output diagnostics in machine-readable JSON format.

```bash
# Example CI environment validation
loy doctor --strict --json
```
