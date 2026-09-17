---
title: "loy make migration"
description: "Scaffold safe database migration with optional zero-downtime recipes"
slug: reference/cli/loy-make-migration
sidebar:
  order: 32
---


Scaffold safe database migration with optional zero-downtime recipes

```
loy make migration <name> [flags]
```

### Options

```
      --column string   Target column for index or shadow-column recipes
  -h, --help            help for migration
      --recipe string   Migration recipe: raw, index-concurrent, or shadow-column (default "raw")
      --table string    Target table for index or column recipes (default "items")
      --type string     Column data type for shadow-column recipe (default "TEXT")
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

