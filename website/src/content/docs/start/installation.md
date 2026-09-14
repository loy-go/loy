---
title: "Installation"
description: "How to install the Loy CLI toolchain on macOS, Linux, and Windows."
---

Loy is distributed as a single static binary with zero external runtime dependencies (`CGO_ENABLED=0`).

## Installation Methods

### Homebrew (macOS & Linux)

The recommended installation channel for macOS and Linux users:

```bash
brew install loy-go/tap/loy
```

Verify the installation:
```bash
loy version
```

---

### Go Toolchain (`go install`)

If you have Go 1.24+ installed on your system:

```bash
go install github.com/loy-go/loy/cmd/loy@latest
```

Ensure your Go binary path (`$GOPATH/bin` or `~/go/bin`) is present in your `$PATH`:

```bash
export PATH="$HOME/go/bin:$PATH"
```

---

### Precompiled Binary Releases

Download precompiled tarballs from the [GitHub Releases](https://github.com/loy-go/loy/releases) page. Every release includes SHA256 checksums and Cosign keyless signatures:

| Platform | Architecture | Binary Format |
|---|---|---|
| **Linux** | `amd64`, `arm64` | `.tar.gz` |
| **macOS** | `amd64` (Intel), `arm64` (Apple Silicon) | `.tar.gz` |
| **Windows** | `amd64`, `arm64` | `.zip` |

#### Linux / macOS Quick Install

```bash
# Example for Linux amd64
curl -sSL https://github.com/loy-go/loy/releases/latest/download/loy_Linux_x86_64.tar.gz | tar -xz -C /usr/local/bin loy
```

---

### Docker Container Image

An official multi-arch container image is published to GitHub Container Registry:

```bash
docker pull ghcr.io/loy-go/loy:latest
docker run --rm -v $(pwd):/workspace -w /workspace ghcr.io/loy-go/loy:latest version
```

---

## Shell Autocompletion

Loy provides shell autocompletion for Bash, Zsh, Fish, and PowerShell:

```bash
# Zsh
echo "autoload -U compinit; compinit" >> ~/.zshrc
echo 'source <(loy completion zsh)' >> ~/.zshrc

# Bash
loy completion bash > /etc/bash_completion.d/loy

# Fish
loy completion fish > ~/.config/fish/completions/loy.fish
```
