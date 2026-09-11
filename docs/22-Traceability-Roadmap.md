# Loy — Requirements Traceability & Final Implementation Roadmap

**Version:** 1.0  
**Status:** Locked  
**Navigation:** [← 21 Reference Application Spec](./21-Reference-Application.md) | [Index](./00-INDEX.md)

---

## Traceability Model

```text
PRD
 ↓
TDS
 ↓
Specification
 ↓
Implementation package
 ↓
Test
 ↓
Acceptance gate
```

## Major Requirement Mapping

| Requirement | Primary Specification | Acceptance |
|---|---|---|
| Project creation | [01-PRD](./01-PRD.md), [11-Project-Workspace](./11-Project-Workspace.md), [10-Configuration-Manifest](./10-Configuration-Manifest.md) | `loy new`, `loy init` |
| Architecture enforcement | [05-Architecture-Rule-Specification](./05-Architecture-Rule-Specification.md), [13-Architecture-Design-Patterns](./13-Architecture-Design-Patterns.md) | `loy check` |
| Deterministic generation | [04-Generator-Specification](./04-Generator-Specification.md), [12-Code-Generation-Templates](./12-Code-Generation-Templates.md) | golden/determinism tests |
| Runtime independence | [08-Runtime-Application-Lifecycle](./08-Runtime-Application-Lifecycle.md) | build/start/shutdown |
| Ecosystem integration | [07-Integration-Capability-Specification](./07-Integration-Capability-Specification.md) | integration tests |
| CLI scripting | [09-CLI-Developer-Experience](./09-CLI-Developer-Experience.md), [14-Error-Diagnostics](./14-Error-Diagnostics.md) | CI acceptance |
| Security | [16-Security](./16-Security.md) | security test suite |
| Observability | [17-Observability](./17-Observability.md) | telemetry/lifecycle tests |
| Deployment | [18-Deployment-Infrastructure](./18-Deployment-Infrastructure.md) | Docker/K8s acceptance |

## Final Implementation Order

1. Foundation
2. Project & Workspace
3. Generator Engine
4. Architecture Engine
5. Core Generators
6. Runtime
7. Integrations
8. CLI/DX completion
9. Fullstack
10. Deployment

## Release Gates

- format
- vet
- unit tests
- generator golden tests
- architecture fixtures
- CLI acceptance
- generated-project acceptance
- build
- runtime lifecycle acceptance
- relevant integrations

## Governance

Future architectural changes must follow:

```text
Idea → ADR → impact analysis → specification update → implementation → acceptance tests
```

## Specification Completion

The Loy specification phase is considered complete and locked for implementation.
