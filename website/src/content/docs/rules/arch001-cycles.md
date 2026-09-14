---
title: "ARCH-001: Dependency Cycles"
description: "Detecting and preventing circular package import chains."
---

## Rule Specification

- **Rule ID**: `ARCH-001`
- **Severity**: `ERROR`
- **Category**: Dependency Graph Integrity

### Rationale
Circular package dependencies make code impossible to decouple, break Go compilation, and indicate severe architectural entanglement.

---

## Violation Example

```text
package a imports package b
package b imports package c
package c imports package a  <-- Cycle detected: a -> b -> c -> a
```

---

## Remediation

1. **Extract Shared Abstraction**: Extract shared domain entities or interfaces into a shared sub-package or the domain layer.
2. **Invert the Dependency**: Pass the required behavior as a consumer-owned interface rather than importing the concrete implementation.
