---
title: "loy migrate"
description: "Manage database schema migrations via embedded goose engine"
slug: reference/cli/loy-migrate
sidebar:
  order: 40
---


Manage database schema migrations via embedded goose engine

### Synopsis

loy migrate provides zero-install database schema evolution powered by an embedded goose migration engine.

### Options

```
      --db-url string      Database connection URL (or DATABASE_URL env var)
      --dir string         Migrations directory (default 'migrations')
      --driver string      Database driver (postgres, default from loy.yaml or postgres)
  -h, --help               help for migrate
      --table string       Goose db version table (default 'goose_db_version')
      --timeout duration   Execution timeout for migration operations (default 30s)
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
* [loy migrate create](/loy/reference/cli/loy-migrate-create/)	 - Create a new timestamped SQL migration file
* [loy migrate down](/loy/reference/cli/loy-migrate-down/)	 - Roll back the latest database migration batch (or to specific version)
* [loy migrate redo](/loy/reference/cli/loy-migrate-redo/)	 - Roll back the most recent migration and re-apply it
* [loy migrate reset](/loy/reference/cli/loy-migrate-reset/)	 - Roll back all database migrations
* [loy migrate status](/loy/reference/cli/loy-migrate-status/)	 - Dump the status of all migrations
* [loy migrate up](/loy/reference/cli/loy-migrate-up/)	 - Apply all pending database migrations
* [loy migrate version](/loy/reference/cli/loy-migrate-version/)	 - Print the current database migration version

