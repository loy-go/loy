---
title: "loy upgrade"
description: "Migrate project configuration and restore managed regions without touching domain logic"
slug: reference/cli/loy-upgrade
sidebar:
  order: 56
---


Migrate project configuration and restore managed regions without touching domain logic

### Synopsis

loy upgrade inspects project configuration (loy.yaml) and managed artifacts (e.g. internal/app/wiring.go),
planning and non-destructively applying schema upgrades and missing managed region markers.
Developer-owned domain and service logic is never rewritten.

```
loy upgrade [path] [flags]
```

### Options

```
  -n, --dry-run   Preview proposed upgrades without modifying files
  -f, --force     Force apply upgrades even if non-critical warnings exist
  -h, --help      help for upgrade
```

### Options inherited from parent commands

```
  -C, --directory string   Change execution directory
      --json               Output results in JSON format
      --no-color           Disable colored ANSI output
  -q, --quiet              Suppress non-essential output
  -v, --verbose            Enable verbose/debug output
```

### SEE ALSO

* [loy](/loy/reference/cli/loy/)	 - Loy — Go developer platform with Laravel-like DX

