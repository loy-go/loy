---
title: "loy make idempotency"
description: "Scaffold request idempotency keys migration and middleware"
slug: reference/cli/loy-make-idempotency
sidebar:
  order: 27
---


Scaffold request idempotency keys migration and middleware

```
loy make idempotency [name] [flags]
```

### Options

```
  -h, --help   help for idempotency
```

### Options inherited from parent commands

```
  -C, --directory string   Change execution directory
      --dry-run            Preview generated operations without writing to disk
      --dual-id            Scaffold dual identifier schema (BIGINT identity + UUID public)
      --force              Overwrite existing files if developer owned
      --json               Output results in JSON format
      --modular            Scaffold sub-domain modular wiring file instead of flat wiring
      --no-color           Disable colored ANSI output
      --no-tidy            Skip running go mod tidy after generation
  -q, --quiet              Suppress non-essential output
      --target string      Target application module in workspace
  -v, --verbose            Enable verbose/debug output
```

### SEE ALSO

* [loy make](/loy/reference/cli/loy-make/)	 - Scaffold application components, slices, and vertical features

