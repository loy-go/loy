# Loy — Error & Diagnostics Specification

**Version:** 1.0  
**Status:** Locked  
**Navigation:** [← 13 Architecture Design Patterns Spec](./13-Architecture-Design-Patterns.md) | [Index](./00-INDEX.md) | [15 Development Workflow Spec →](./15-Development-Workflow-Toolchain.md)

---

## Diagnostic Model

```go
type Diagnostic struct {
    Severity Severity
    Code     string
    Message  string
    Detail   string
    Hint     string
    File     string
    Line     int
    Column   int
}
```

## Severity

`info`, `warning`, `error`.

## Code Namespaces

CLI, configuration, generator, architecture, integration, project, workspace, security, development and deployment namespaces use stable codes. Architecture rules retain stable `ARCH-*` identifiers.

## Error Quality

A diagnostic should answer: what happened, where, why, and how to fix it.

## Rendering

Human and JSON output are supported. JSON must remain machine-parseable on failure.

## Internal Errors

Unexpected internal failures use an internal-error category such as `LOY999` and should provide a debug path without exposing secrets.

---

**Next:** [15-Development-Workflow-Toolchain.md — Development Workflow & Toolchain Specification](./15-Development-Workflow-Toolchain.md)
