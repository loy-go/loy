---
title: "loy make view"
description: "Scaffold Templ view component or page"
slug: reference/cli/loy-make-view
sidebar:
  order: 48
---


Scaffold Templ view component or page

```
loy make view <name> [flags]
```

### Options

```
  -h, --help      help for view
      --partial   Scaffold as partial UI component instead of full page
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

