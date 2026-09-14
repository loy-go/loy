---
title: "Docker, Kubernetes & Deployment"
description: "Deploying Loy applications to production with multi-stage Dockerfiles, Kubernetes manifests, and Helm."
---

Loy provides built-in deployment scaffolding through `loy make deploy` or individual deployment generators.

## 1. Multi-Stage Dockerfile Scaffolding

Generate a production-ready, minimal, non-root Dockerfile and local `docker-compose.yml`:

```bash
loy make docker
```

### Generated Dockerfile Architecture

- **Stage 1 (Builder)**: Uses `golang:1.24-alpine` with Go module cache mounts (`--mount=type=cache,target=/go/pkg/mod`).
- **Static Compilation**: Compiles with `CGO_ENABLED=0 go build -trimpath -ldflags="-s -w"`.
- **Stage 2 (Runtime)**: Runs on minimal Distroless or Alpine runtime as non-root user (`USER 10001:10001`) with CA certificates and timezone data. Resulting container size is typically < 25MB.

---

## 2. Kubernetes Manifests

Generate cloud-native Kubernetes manifests:

```bash
loy make k8s
```

Generated resources in `deploy/k8s/`:
- **`deployment.yaml`**: Configures replica count, non-root security context, CPU/memory resource requests/limits, and liveness/readiness probes targeting `/health/live` and `/health/ready`.
- **`service.yaml`**: ClusterIP service routing traffic to port 8080.
- **`ingress.yaml`**: Ingress rules with TLS configuration.
- **`configmap.yaml`**: Non-sensitive environment variables.

---

## 3. Helm Charts

Generate a customizable Helm chart for enterprise package distribution:

```bash
loy make helm
```

---

## 4. Continuous Integration Pipelines

Generate automated GitHub Actions or GitLab CI workflows:

```bash
loy make ci --provider=github
```
