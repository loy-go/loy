# Loy — Versioning, Upgrade & Migration Specification

**Version:** 1.0  
**Status:** Locked  
**Navigation:** [← 19 Plugin Extension Spec](./19-Plugin-Extension.md) | [Index](./00-INDEX.md) | [21 Reference Application Spec →](./21-Reference-Application.md)

---

## Versioning

Loy follows semantic versioning.

## Breaking Interfaces

The following can be breaking changes:

- commands and flags
- JSON output schemas
- architecture rules
- generation output
- configuration schema
- plugin APIs

## Configuration Version

`loy.yaml` has an independent schema version.

## Generated Code

Generated output is part of Loy's effective public interface. Generator output changes may therefore constitute breaking changes.

## Upgrade

`loy upgrade` may migrate configuration and assist with managed generated artifacts, but must not silently rewrite developer-owned code.

## Migration Safety

Migration should use plan → preview → apply where practical. Dry-run is preferred for potentially broad changes.

## Dependency Upgrades

Loy version changes must not silently upgrade application dependencies.

---

**Next:** [21-Reference-Application.md — Reference Application Specification](./21-Reference-Application.md)
