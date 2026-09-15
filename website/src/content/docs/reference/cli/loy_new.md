---
title: "loy new"
description: "Create a new Loy project with preset configuration"
slug: reference/cli/loy-new
sidebar:
  order: 45
---


Create a new Loy project with preset configuration

```
loy new [project-name] [flags]
```

### Options

```
      --cache string          Cache adapter (valkey, redis, memory, none)
      --db string             Database adapter (postgres, sqlite, mysql, none)
  -f, --force                 Overwrite destination directory if exists
  -h, --help                  help for new
      --http string           HTTP adapter (fiber, chi, gin, nethttp, echo)
  -i, --interactive           Interactive project setup wizard
      --multi-tenant string   Multi-tenancy strategy (rls, column)
  -p, --preset string         Preset template (api, fullstack, minimal, monorepo, saas, web) (default "api")
      --queue string          Queue adapter (asynq, river, none)
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

