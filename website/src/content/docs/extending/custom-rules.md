---
title: "Extending Loy: Custom Architecture Rules"
description: "How to define organizational architecture rules and custom scaffolding templates."
---

Engineering organizations can customize Loy's code generators and static architecture validation to enforce internal standards.

## 1. Customizing Template Overrides

Place custom template files in `.loy/templates/` to override default generators:

```
.loy/
└── templates/
    └── handler.go.tmpl       # Custom company handler template
```

When running `loy make handler` or `loy make crud`, Loy prefers workspace-local templates over built-in defaults.

---

## 2. Defining Custom Package Layer Overrides

If your organization uses specialized package directories (e.g. `internal/integrations/stripe`), define layer mappings in `loy.yaml`:

```yaml title="loy.yaml"
architecture:
  strict: true
  overrides:
    "github.com/myorg/app/internal/integrations/stripe": "Infrastructure"
    "github.com/myorg/app/internal/graphql": "Transport"
```

---

## 3. Pre-Commit Architecture Enforcement

Ensure all team members validate architecture before committing by installing the pre-commit hook:

```bash
loy hook install
```

This creates `.git/hooks/pre-commit`, blocking commits containing layer violations locally in less than 50 milliseconds.
