# Loy Implementation Plans Index

This directory contains the standard engineering pipeline and detailed implementation plans for each of the 10 engineering phases defined in [03-TIP.md](../03-TIP.md).

## Implementation Standard

- **[IMPLEMENTATION-PIPELINE.md](./IMPLEMENTATION-PIPELINE.md)** — Standard 6-step engineering pipeline & verification loop established from Phase 1 learnings.

## Phase Roadmap

| Phase | Plan Document | Primary Scope | Status |
|---|---|---|---|
| **01** | [01-Phase-Foundation.md](./01-Phase-Foundation.md) | CLI root, OS filesystem abstraction, process runner, diagnostics model | Completed |
| **02** | [02-Phase-Project-System.md](./02-Phase-Project-System.md) | Discovery, `loy.yaml` manifest, `go.work` resolution, presets, `loy new` | Completed |
| **03** | [03-Phase-Generator-Engine.md](./03-Phase-Generator-Engine.md) | Generator contracts, naming rules, `text/template` + `gofmt`, plans, splicing | Completed |
| **04** | [04-Phase-Architecture-Engine.md](./04-Phase-Architecture-Engine.md) | Two-phase analysis, dependency graphs, cycle detection, 14 rules, suppressions | Completed |
| **05** | [05-Phase-Core-Generators.md](./05-Phase-Core-Generators.md) | `make model`, `repo`, `svc`, `handler`, `feature`, `crud`, managed wiring | Completed |
| **06** | [06-Phase-Runtime.md](./06-Phase-Runtime.md) | Composition root, lifecycle states, signal handling, graceful shutdown, health | Completed |
| **07** | [07-Phase-Integrations.md](./07-Phase-Integrations.md) | Fiber, embedded `goose` migrations, `sqlc`, Valkey, Asynq, OpenTelemetry | Completed |
| **08** | [08-Phase-Developer-Experience.md](./08-Phase-Developer-Experience.md) | Native `loy dev` supervisor, `loy doctor`, `loy graph`, shell completion | Ready |
| **09** | [09-Phase-Fullstack.md](./09-Phase-Fullstack.md) | Templ SSR, HTMX patterns, Vite asset pipeline integration | Ready |
| **10** | [10-Phase-Deployment.md](./10-Phase-Deployment.md) | Multi-stage Dockerfiles, Kubernetes manifests, Helm charts, CI pipelines | Ready |

---

[Back to Documentation Index](../00-INDEX.md)
