# Loy — Test Strategy & Acceptance Specification

**Version:** 1.0  
**Status:** Locked  
**Navigation:** [← 05 Architecture Rule Spec](./05-Architecture-Rule-Specification.md) | [Index](./00-INDEX.md) | [07 Integration Capability Spec →](./07-Integration-Capability-Specification.md)

---

## Test Layers

1. Unit
2. Golden
3. Component
4. Integration
5. CLI
6. Architecture fixtures
7. Generated project
8. Workspace
9. E2E
10. Regression

## Critical Paths

Project discovery, manifest parsing, generator plans, filesystem safety, conflict detection, source template rendering, architecture graph/rules, CLI execution and generated application compilation.

## Golden Tests

Generator outputs and diagnostic output may use golden fixtures. Updates must be explicit and reviewable.

## Generated Project Acceptance

```text
loy new testapp
cd testapp
loy make feature users
loy check
go test ./...
go build ./...
```

## Runtime Acceptance

Generate → build → start → readiness → request → SIGTERM → clean exit.

## Race/Static Checks

Concurrency-sensitive code should run with the race detector. CI should run `go vet ./...`.

## Determinism

The same generator input/configuration must produce byte-equivalent managed output.

## CI Gates

Formatting, vet, unit tests, golden tests, architecture checks, CLI acceptance, generated project acceptance, and relevant integration tests are release gates.

## Coverage

Coverage is a signal, not a target. Critical behavior receives tests even when line coverage is already high elsewhere.

---

**Next:** [07-Integration-Capability-Specification.md — Integration & Capability Specification](./07-Integration-Capability-Specification.md)
