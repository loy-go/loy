---
title: "Architecture Drift & Git Diffing"
description: "Detecting architectural erosion and package dependency drift between Git branches using loy graph --diff."
---

Software architectures rarely collapse in a single commit; they erode through small, gradual boundary violations merged over months—a transport utility imported into domain here, an un-inverted SQL repository imported there.

Loy stops architectural drift before code merges by comparing package graphs directly against Git references:
```bash
loy graph --diff <base-ref> [--format ascii|mermaid|markdown]
```

---

## Comparing Against Base Branches

You can compare your current working tree against any Git branch, commit SHA, or tag:

```bash
# Compare against main
loy graph --diff main

# Compare against remote tracking branch
loy graph --diff origin/main

# Compare against previous commit
loy graph --diff HEAD~1
```

### How It Works Internally

1. **Base Extraction**: Loy invokes `git archive --format=tar <base-ref>` to unpack Go source files, `go.mod`, and `loy.yaml` into an isolated, temporary sandboxed directory.
2. **Dual-Model Graphing**: Runs Loy's Phase 1 AST analyzer on both the extracted base and current working tree.
3. **Delta Computation**: Identifies added/removed packages, new/removed dependencies, and newly introduced vs. resolved rule violations.
4. **Deterministic Output**: Formats the comparison report in your requested format without writing temporary files to disk.

---

## Output Formats

### 1. ASCII Terminal View (Default)
```text
Architecture Drift (compared to main):

[!] NEW ARCHITECTURAL VIOLATIONS:
  ! github.com/example/app/internal/order/domain -> github.com/example/app/internal/platform/database

[+] Added Packages:
  + github.com/example/app/internal/invoice/domain [Domain]
  + github.com/example/app/internal/invoice/service [Application]

[+] Added Dependencies:
  + github.com/example/app/internal/invoice/service -> github.com/example/app/internal/invoice/domain
```

### 2. Mermaid Highlighted Diagram (`--format mermaid`)
Renders an interactive Mermaid diagram highlighting added edges in solid blue and new architectural violations in bold red.

### 3. Markdown PR Comment Report (`--format markdown`)
Formatted specifically for GitHub Actions or GitLab CI pull request comments:

```markdown
### 🏛️ Architecture Drift Report

**Comparison Base:** `main`

🚨 **1 New Architectural Violation(s) Introduced!**

| From Package | To Package | Status |
|---|---|---|
| `internal/order/domain` | `internal/platform/database` | ❌ VIOLATION |

<details><summary>📦 Package Changes</summary>
**Added:**
- `+ internal/invoice/domain` (Domain)
- `+ internal/invoice/service` (Application)
</details>
```

---

## GitHub Actions CI Integration

Post automated architectural drift reviews directly into Pull Requests:

```yaml
name: Architecture Drift Guard

on:
  pull_request:
    branches: [main]

jobs:
  arch-diff:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0 # fetch all history for diffing

      - uses: actions/setup-go@v5
        with:
          go-version: '1.24'

      - name: Install Loy
        run: go install github.com/loy-go/loy/cmd/loy@latest

      - name: Generate Drift Report
        run: |
          loy graph --diff origin/main --format markdown > drift-report.md
          cat drift-report.md

      - name: Enforce Clean Architecture
        run: loy check --strict
```
