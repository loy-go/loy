---
title: "Universal Agent Rules Generator"
description: "Turnkey AI agent instruction scaffolding for Cursor, Claude Code, GitHub Copilot, and Windsurf via loy make agent-rules."
---

Different AI coding tools read project rules from different metadata files. Maintaining these instruction prompts manually leads to configuration drift, missed boundary checks, and accidental file corruptions.

Loy provides a single turnkey command:
```bash
loy make agent-rules [--target all|cursor|claude|copilot|windsurf] [--force]
```

---

## Scaffolding All Assistants

Running `loy make agent-rules` generates authoritative instruction prompts tailored to every major AI toolchain:

```bash
loy make agent-rules
```

### Generated Artifacts

| Assistant | Target File | Core Enforcements |
|---|---|---|
| **Claude Code / Generic LLMs** | `AGENTS.md` & `CLAUDE.md` | Strict 4-layer Clean Architecture invariants, verification commands (`loy check --format agent`, `make check`), comment region rules. |
| **Cursor IDE** | `.cursor/rules/loy.mdc` & `.cursorrules` | Architectural boundary enforcement on every edit, prohibited import rules, self-healing diagnostics. |
| **GitHub Copilot** | `.github/copilot-instructions.md` | Contextual workspace instructions enforcing domain purity and explicit constructor wiring. |
| **Windsurf** | `.windsurfrules` | Cascade AI instructions for Clean Architecture layer flow and comment region preservation. |

---

## Targeting a Specific Assistant

If your team exclusively uses one tool, use `--target` (or `--for` / `--client`):

```bash
# Generate only Cursor IDE rules
loy make agent-rules --target cursor

# Generate only Claude Code instructions
loy make agent-rules --target claude

# Generate only GitHub Copilot instructions
loy make agent-rules --target copilot

# Generate only Windsurf rules
loy make agent-rules --target windsurf
```

---

## What the Rules Instruct AI Agents

The generated rule files provide external agents with complete architectural boundaries:

1. **Layer Hierarchy**:
   - `Transport -> Application -> Domain <- Infrastructure`
   - Domain is pure Go (no `database/sql`, ORMs, or web frameworks).
   - Application depends on domain interfaces, not concrete adapters.
   - Handlers in Transport delegate all mutations to Application services.
2. **Guarded Comment Regions**:
   - External agents are forbidden from modifying code outside `// loy:region:<name>` and `// loy:endregion` blocks in managed files.
   - Agents are explicitly instructed never to delete or nest region boundary markers.
3. **Explicit Dependency Injection**:
   - Agents are warned against using service locators or dynamic reflection containers (`uber/dig`).
   - All wiring must occur in `internal/app/wiring.go`.
4. **Mandatory Verification Gate**:
   - Agents are instructed to run `loy check --format agent` and `go test -v -race ./...` before considering any task complete.

---

## Overwriting Existing Files

Because rule documents are considered developer-owned once generated, re-generating rules with modifications requires the `--force` flag:

```bash
loy make agent-rules --force
```

Use `--dry-run` to preview planned operations without modifying your filesystem:
```bash
loy make agent-rules --dry-run
```
