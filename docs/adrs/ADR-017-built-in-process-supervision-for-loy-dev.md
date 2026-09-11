# ADR-017: Built-in Process Supervision for loy dev

## Status
Accepted

## Context
Fullstack and service development requires running multiple concurrent tasks (HTTP server, background queue workers, Templ compiler, Vite asset bundler). Shelling out to third-party tools like `air` complicates toolchains and struggles to cleanly coordinate multi-process lifecycles.

## Decision
1. Implement native file-watching and process supervision directly inside `loy dev` using `fsnotify`.
2. Supervise target processes (API, worker, frontend build) with process tree tracking.
3. On SIGINT/SIGTERM or file change, send graceful shutdown signals to all child process trees, with forced SIGKILL only after a configured grace timeout.
4. Suppress child orphan processes on abrupt exit.

## Consequences
- Zero external runner dependency for development mode.
- Uniform logging prefixing and status reporting in developer terminal.
- Reliable lifecycle shutdown of all dependent dev daemons.

---

[Back to ADR Index](./README.md) | [Back to Documentation Index](../00-INDEX.md)
