# Loy — Security Specification

**Version:** 1.0  
**Status:** Locked  
**Navigation:** [← 15 Development Workflow Spec](./15-Development-Workflow-Toolchain.md) | [Index](./00-INDEX.md) | [17 Observability Spec →](./17-Observability.md)

---

## CLI Security

Protect against path traversal, command injection, unsafe file writes, secret leakage and malicious configuration.

## Process Execution

External processes are invoked directly; user input is never interpolated into shell commands.

## File Safety

Generated paths must be normalized, relative to the intended project root and incapable of escaping it.

## Secrets

Credentials are external to source configuration. Environment variables, secret managers and deployment platform secrets are preferred.

## Logging

Secrets must never appear in normal, verbose, debug, doctor, JSON or stack-trace output.

## Generated Application Security

Security-sensitive templates should provide sensible conventions for password hashing, cookies, CSRF, CORS, request limits, TLS, authentication, authorization and secret handling.

## Supply Chain

Future plugin support must define origin, integrity, permissions, compatibility and trust before becoming stable.

---

**Next:** [17-Observability.md — Observability Specification](./17-Observability.md)
