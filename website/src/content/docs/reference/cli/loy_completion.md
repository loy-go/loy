---
title: "loy completion"
description: "Generate shell completion scripts"
slug: reference/cli/loy-completion
sidebar:
  order: 2
---


Generate shell completion scripts

### Synopsis

Generate shell completion script for loy.
To load completions:

Bash:
  $ source <(loy completion bash)

Zsh:
  # If shell completion is not already enabled in your environment:
  $ echo "autoload -U compinit; compinit" >> ~/.zshrc
  $ source <(loy completion zsh)

Fish:
  $ loy completion fish | source

PowerShell:
  PS> loy completion powershell | Out-String | Invoke-Expression


```
loy completion [bash|zsh|fish|powershell]
```

### Options

```
  -h, --help   help for completion
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

* [loy](/reference/cli/loy/)	 - Loy — Go developer platform with Laravel-like DX

