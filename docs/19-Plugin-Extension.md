# Loy — Plugin & Extension Specification

**Version:** 1.0  
**Status:** Post-MVP boundary locked  
**Navigation:** [← 18 Deployment Infrastructure Spec](./18-Deployment-Infrastructure.md) | [Index](./00-INDEX.md) | [20 Versioning Upgrade Spec →](./20-Versioning-Upgrade-Migration.md)

---

## Status

Plugin architecture is not an MVP implementation requirement.

## Candidate Extension Points

- commands
- generators
- presets
- integrations
- diagnostics

## Design Rule

Stabilize internal interfaces first. Do not publish an external plugin API until real use cases prove the abstraction.

## Future Security Requirements

Origin, integrity, compatibility, permissions, loading, execution isolation and failure behavior must be defined before plugins become stable.

## Prohibitions

Plugins must not force reflection DI, global service locators or runtime component discovery into the core framework.

---

**Next:** [20-Versioning-Upgrade-Migration.md — Versioning, Upgrade & Migration Specification](./20-Versioning-Upgrade-Migration.md)
