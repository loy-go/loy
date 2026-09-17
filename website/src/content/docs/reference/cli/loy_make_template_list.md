---
title: "loy make template list"
description: "List all built-in templates available for override"
slug: reference/cli/loy-make-template-list
sidebar:
  order: 45
---


List all built-in templates available for override

```
loy make template list [flags]
```

### Options

```
  -h, --help   help for list
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

* [loy make template](/loy/reference/cli/loy-make-template/)	 - Manage and eject generator templates for local project customization

