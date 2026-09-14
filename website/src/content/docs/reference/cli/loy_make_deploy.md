---
title: "loy make deploy"
description: "Scaffold production deployment assets (docker, k8s, helm, ci, or all)"
slug: reference/cli/loy-make-deploy
sidebar:
  order: 10
---


Scaffold production deployment assets (docker, k8s, helm, ci, or all)

### Synopsis

Scaffold deployment and infrastructure assets:
  - docker: Multi-stage Dockerfile & docker-compose.yml
  - k8s:    Kubernetes manifests in deploy/k8s/
  - helm:   Helm chart in deploy/helm/<app>/
  - ci:     CI pipeline workflow (.github/workflows/ci.yml)
  - all:    Scaffold all deployment targets together (default)

```
loy make deploy [type] [flags]
```

### Options

```
  -h, --help                   help for deploy
      --provider string        CI provider (github or gitlab) (default "github")
      --target-binary string   Target command binary to compile for Dockerfile
```

### Options inherited from parent commands

```
  -C, --directory string   Change execution directory
      --dry-run            Preview generated operations without writing to disk
      --force              Overwrite existing files if developer owned
      --json               Output results in JSON format
      --no-color           Disable colored ANSI output
  -q, --quiet              Suppress non-essential output
      --target string      Target application module in workspace
  -v, --verbose            Enable verbose/debug output
```

### SEE ALSO

* [loy make](/reference/cli/loy-make/)	 - Scaffold application components, slices, and vertical features

