# ADR-014: Managed Code Splicing via Comment Regions

## Status
Accepted

## Context
When running commands like `loy make feature <name>` or `loy make crud <name>`, new services, repositories, and handlers need to be registered in the application composition root (`internal/app/wiring.go`). Full AST rewriting (`go/ast`) is brittle when developer-owned manual edits exist. Manual paste output harms developer experience.

## Decision
1. Implement managed file regions via structured boundary comments:
   `// loy:region:<section_name>`
   `// loy:endregion`
2. Generator identifies the targeted region, splices structured registration statements (e.g. self-registering feature slices), and runs `gofmt` over the resulting file.
3. If boundary comments are missing or corrupted, `loy make` fails with a clear diagnostic (`LOY-GEN-002`) directing the developer to either restore the comment tag or wire manually.

## Consequences
- Deterministic, safe updates to developer-owned composition files.
- Transparent and human-readable boundary markers.
- Survives arbitrary developer edits outside managed comments.

---

[Back to ADR Index](./README.md) | [Back to Documentation Index](../00-INDEX.md)
