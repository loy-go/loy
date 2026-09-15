---
title: "Community Plugins & Custom Generators"
description: "Extending Loy with sandboxed out-of-process community plugins and custom generators via loy plugin."
---

While Loy provides built-in generators for models, repositories, services, handlers, gRPC, and WebSockets, organizations often require domain-specific generators (e.g. Apollo GraphQL schemas, AWS CDK constructs, Kafka consumers, or Stripe webhook handlers).

Loy provides an out-of-process, sandboxed plugin ecosystem via:
```bash
loy plugin <command> [flags]
```

---

## Installing Plugins

Install any community plugin directly from a Git repository:

```bash
loy plugin install github.com/loy-community/graphql-generator
```

During installation, Loy:
1. Performs a shallow Git clone into your local plugin cache (`~/.config/loy/plugins/` or `.loy/plugins/`).
2. Validates the `loy-plugin.yaml` manifest schema.
3. Checks executable permissions and performs path traversal jail validation.

---

## Listing Installed Plugins

View all active plugins and their metadata:

```bash
loy plugin list
```

Output:
```text
NAME                VERSION   DESCRIPTION
graphql-generator   v1.2.0    Scaffold Apollo-compatible GraphQL resolvers and schemas
kafka-consumer      v0.4.1    Scaffold Kafka event consumers with Sarama adapters
```

---

## Executing a Plugin

Invoke an installed plugin and pass generator arguments:

```bash
loy plugin run graphql-generator order id:string total:float
```

### Sandboxed Subprocess Security

Loy guarantees safe plugin execution:
- **Zero Shell Interpolation**: Subprocesses are invoked directly with argument slices (`exec.CommandContext`), preventing shell injection attacks.
- **Path Sandboxing**: Plugins cannot write outside the resolved target project root directory.
- **Controlled Environment**: Environment variables are scrubbed; secrets are never passed automatically to community binaries.

---

## Authoring a Loy Plugin

Any executable binary or script (Go, Bash, Python, Node.js) can become a Loy plugin. Simply place a `loy-plugin.yaml` manifest in your repository root:

```yaml
version: 1
name: my-generator
description: "Scaffold internal enterprise service templates"
executable: "./bin/generator"
author: "Engineering Platform Team"
hooks:
  post-generate: "go mod tidy"
```

When invoked via `loy plugin run my-generator`, your executable receives user arguments on `os.Args[1:]` and writes planned output into the target working directory.
