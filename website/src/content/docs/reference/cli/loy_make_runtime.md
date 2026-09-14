---
title: "loy make runtime"
description: "Scaffold application runtime lifecycle, composition root, and health checks"
slug: reference/cli/loy-make-runtime
sidebar:
  order: 30
---


Scaffold application runtime lifecycle, composition root, and health checks

```
loy make runtime [flags]
```

### Options

```
  -h, --help          help for runtime
      --http string   HTTP framework to scaffold (fiber or nethttp)
```

### Options inherited from parent commands

```
  -C, --directory string   Change execution directory
      --dry-run            Preview generated operations without writing to disk
      --force              Overwrite existing files if developer owned
      --json               Output results in JSON format
      --modular            Scaffold sub-domain modular wiring file instead of flat wiring
      --no-color           Disable colored ANSI output
  -q, --quiet              Suppress non-essential output
      --target string      Target application module in workspace
  -v, --verbose            Enable verbose/debug output
```

### SEE ALSO

* [loy make](/loy/reference/cli/loy-make/)	 - Scaffold application components, slices, and vertical features

