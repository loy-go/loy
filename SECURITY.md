# Security Policy

The Loy project takes security seriously. We appreciate responsible disclosure from researchers and users to protect the entire community.

## Supported Versions

We provide security patches and vulnerability updates for the following versions of Loy:

| Version | Supported          |
| ------- | ------------------ |
| 1.x     | :white_check_mark: |
| < 1.0   | :white_check_mark: (latest minor) |

## Reporting a Vulnerability

**Please DO NOT report security vulnerabilities through public GitHub issues, discussions, or social media channels.**

To report a vulnerability:
1. **GitHub Private Vulnerability Reporting (Preferred)**:
   Navigate to the repository's **Security** tab and click **Report a vulnerability** to open a private draft advisory.
2. **Email**:
   If private vulnerability reporting is unavailable, email details to **[security@loy.dev](mailto:security@loy.dev)** with:
   - Type of vulnerability (e.g., path traversal, code injection, denial of service).
   - Step-by-step instructions or proof-of-concept to reproduce the issue.
   - Affected CLI commands, generators, or architectural rules.
   - Any proposed remediation or patches.

## Response SLA

- **Initial Response**: Within 48 hours of receipt.
- **Triage & Reproduction**: Within 5 business days.
- **Remediation & Patch Release**: Critical vulnerabilities patched within 14 business days; advisory published via GitHub Security Advisories and CVE identifier assigned.

## Security Practices in Loy

- **Path Sandboxing**: All generator file mutations are validated via `filesystem.CleanAndValidatePath` to prevent directory traversal outside project boundaries.
- **Zero Shell Interpolation**: Subprocesses are invoked directly with distinct argument slices via `os/exec.CommandContext`, strictly avoiding `sh -c` or shell strings.
- **Dependency Scanning**: `govulncheck` is executed in our continuous integration pipeline to prevent known vulnerabilities in the dependency graph.
