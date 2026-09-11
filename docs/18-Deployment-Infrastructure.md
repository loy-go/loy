# Loy — Deployment & Infrastructure Specification

**Version:** 1.0  
**Status:** Locked  
**Navigation:** [← 17 Observability Spec](./17-Observability.md) | [Index](./00-INDEX.md) | [19 Plugin Extension Spec →](./19-Plugin-Extension.md)

---

## Principle

Deployment assets are generated infrastructure. Running applications do not depend on Loy.

## Docker

Production images should use multi-stage builds where appropriate and execute the compiled application binary directly.

## Runtime Configuration

Use environment/deployment configuration and secret managers rather than `loy.yaml`.

## Kubernetes

Optional generated resources may include Deployments, Services, ConfigMaps, secret references, probes, Ingress and autoscaling where configured.

## Helm

Optional packaging integration for Kubernetes. It is not a requirement for Kubernetes deployment.

## Health

Map liveness/readiness probes to the application's documented health endpoints.

## Shutdown

Applications must handle SIGTERM and complete graceful shutdown within the configured termination budget.

---

**Next:** [19-Plugin-Extension.md — Plugin & Extension Specification](./19-Plugin-Extension.md)
