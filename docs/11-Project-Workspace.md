# Loy — Project & Workspace System Specification

**Version:** 1.0  
**Status:** Locked  
**Navigation:** [← 10 Configuration Manifest Spec](./10-Configuration-Manifest.md) | [Index](./00-INDEX.md) | [12 Code Generation Templates Spec →](./12-Code-Generation-Templates.md)

---

## Project

A project is the primary Loy-managed unit and normally owns one authoritative `go.mod` and `loy.yaml`.

## Workspace

A workspace can contain multiple projects/modules and is represented by `go.work` for Go semantics.

Example:

```text
workspace/
├── go.work
├── loy.yaml
├── apps/
│   ├── api/
│   └── worker/
└── packages/
    ├── auth/
    └── observability/
```

## Apps and Packages

Apps are runnable projects. Packages are reusable modules. By default:

```text
App → Package     allowed
Package → Package allowed
App → App         disallowed
Package → App     disallowed
```

## Targets

Targets represent runnable/operational units such as `api`, `worker`, or `scheduler`.

## Discovery

Use explicit directory first, then current directory, then parents, then workspace hints. `go env GOMOD` and `go env GOWORK` may be used to resolve authoritative Go locations.

## Graphs

Loy can model module, package, workspace and architecture graphs. These are distinct semantics even if they share graph infrastructure.

## Workspace Acceptance

Workspace projects must build, test, resolve dependencies and pass architecture validation under documented target selection rules.

---

**Related ADRs:**
- [ADR-008: Go Workspace Authority](./adrs/ADR-008-go-workspace-authority.md)

**Next:** [12-Code-Generation-Templates.md — Code Generation & Template Specification](./12-Code-Generation-Templates.md)
