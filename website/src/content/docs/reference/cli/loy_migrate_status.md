---
title: "loy migrate status"
description: "Dump the status of all migrations"
slug: reference/cli/loy-migrate-status
sidebar:
  order: 41
---


Dump the status of all migrations

```
loy migrate status [flags]
```

### Options

```
  -h, --help   help for status
```

### Options inherited from parent commands

```
      --db-url string      Database connection URL (or DATABASE_URL env var)
      --dir string         Migrations directory (default 'migrations')
  -C, --directory string   Change execution directory
      --driver string      Database driver (postgres, default from loy.yaml or postgres)
      --json               Output results in JSON format
      --no-color           Disable colored ANSI output
  -q, --quiet              Suppress non-essential output
      --table string       Goose db version table (default 'goose_db_version')
      --timeout duration   Execution timeout for migration operations (default 30s)
  -v, --verbose            Enable verbose/debug output
```

### SEE ALSO

* [loy migrate](/loy/reference/cli/loy-migrate/)	 - Manage database schema migrations via embedded goose engine

