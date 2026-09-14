---
title: "loy make docker"
description: "Scaffold production multi-stage Dockerfile and docker-compose.yml"
slug: reference/cli/loy-make-docker
sidebar:
  order: 11
---


Scaffold production multi-stage Dockerfile and docker-compose.yml

```
loy make docker [name] [flags]
```

### Options

```
  -h, --help                   help for docker
      --target-binary string   Target command binary to compile (e.g. api, web, worker)
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

* [loy make](/loy/reference/cli/loy-make/)	 - Scaffold application components, slices, and vertical features

