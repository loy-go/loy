# Loy Implementation Plans Index

This directory contains the detailed, comprehensive implementation plans for each of the 10 engineering phases defined in [03-TIP.md](../03-TIP.md).

## Phase Roadmap

| Phase | Plan Document | Primary Scope | Status |
|---|---|---|---|
| **01** | [01-Phase-Foundation.md](./01-Phase-Foundation.md) | CLI root, OS filesystem abstraction, process runner, diagnostics model | Ready |
| **02** | [02-Phase-Project-System.md](./02-Phase-Project-System.md) | Discovery, `loy.yaml` manifest, `go.work` resolution, presets, `loy new` | Ready |
| **03** | [03-Phase-Generator-Engine.md](./03-Phase-Generator-Engine.md) | Generator contracts, naming rules, `text/template` + `gofmt`, plans, splicing | Ready |
| **04** | [04-Phase-Architecture-Engine.md](./04-Phase-Architecture-Engine.md) | Two-phase analysis, dependency graphs, cycle detection, 14 rules, suppressions | Ready |
| **05** | [05-Phase-Core-Generators.md](./05-Phase-Core-Generators.md) | `make model`, `repo`, `svc`, `handler`, `feature`, `crud`, managed wiring | Ready |
| **06** | [06-Phase-Runtime.md](./06-Phase-Runtime.md) | Composition root, lifecycle states, signal handling, graceful shutdown, health | Ready |
| **07** | [07-Phase-Integrations.md](./07-Phase-Integrations.md) | Fiber, embedded `goose` migrations, `sqlc`, Valkey, Asynq, OpenTelemetry | Ready |
| **08** | [08-Phase-Developer-Experience.md](./08-Phase-Developer-Experience.md) | Native `loy dev` supervisor, `loy doctor`, `loy graph`, shell completion | Ready |
| **09** | [09-Phase-Fullstack.md](./09-Phase-Fullstack.md) | Templ SSR, HTMX patterns, Vite asset pipeline integration | Ready |
| **10** | [10-Phase-Deployment.md](./10-Phase-Deployment.md) | Multi-stage Dockerfiles, Kubernetes manifests, Helm charts, CI pipelines | Ready |

---

[Back to Documentation Index](../00-INDEX.md)
