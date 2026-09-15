---
title: "loy seed"
description: "Execute database seeders to populate development fixtures"
slug: reference/cli/loy-seed
sidebar:
  order: 54
---


Execute database seeders to populate development fixtures

### Synopsis

loy seed runs registered database seeders inside transactions to populate initial data.

```
loy seed [path] [flags]
```

### Options

```
      --fake   Generate randomized fake data using gofakeit
  -h, --help   help for seed
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

