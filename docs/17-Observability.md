# Loy — Observability Specification

**Version:** 1.0  
**Status:** Locked  
**Navigation:** [← 16 Security Spec](./16-Security.md) | [Index](./00-INDEX.md) | [18 Deployment Infrastructure Spec →](./18-Deployment-Infrastructure.md)

---

## Logging

Use Go `log/slog` as the default structured logging foundation.

## OpenTelemetry

Preferred tracing/metrics integration. Generated applications should support initialization and clean shutdown.

## Instrumentation Boundaries

HTTP, gRPC, database, queue, external HTTP and background jobs are natural instrumentation boundaries.

## Context

Trace/request context should flow transport → application → infrastructure and be propagated to background work where appropriate.

## Vendor Neutrality

Domain code must not depend on OpenTelemetry or vendor SDKs.

## Health

HTTP services should provide `/health/live` and `/health/ready` with liveness independent from full dependency health.

---

**Next:** [18-Deployment-Infrastructure.md — Deployment & Infrastructure Specification](./18-Deployment-Infrastructure.md)
