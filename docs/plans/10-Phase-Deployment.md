# Phase 10: Deployment & Infrastructure Implementation Plan

**Phase:** 10 of 10  
**Status:** Ready for Implementation  
**Estimated Scope:** Multi-stage Dockerfiles, Kubernetes manifests, Helm charts, CI pipeline templates  
**Primary Specifications:** [18-Deployment-Infrastructure.md](../18-Deployment-Infrastructure.md), [16-Security.md](../16-Security.md), [22-Traceability-Roadmap.md](../22-Traceability-Roadmap.md)

---

## 1. Goal & Objectives
Provide production readiness and artifact distribution tooling:
- Multi-stage `Dockerfile` generating minimal, secure, non-root container images.
- Cloud-native Kubernetes manifests (Deployments, Services, ConfigMaps, Secret references, Probes).
- Optional Helm chart generation.
- CI/CD workflow generators for GitHub Actions and GitLab CI.
- Verification of the canonical end-to-end acceptance loop.

---

## 2. Package Architecture & Artifact Templates

```text
generators/
└── deployment/
    ├── docker/
    │   ├── dockerfile.go   # Multi-stage build generator
    │   └── compose.go      # docker-compose.yml for local PostgreSQL/Valkey dev
    ├── k8s/
    │   ├── deployment.go   # K8s Deployment with health probes & securityContext
    │   ├── service.go      # ClusterIP / Ingress resource
    │   └── configmap.go    # Non-sensitive runtime configuration
    ├── helm/
    │   └── chart.go        # Parameterized values.yaml and templates/
    └── ci/
        └── github.go       # .github/workflows/ci.yml (format, vet, test, check, build)
```

---

## 3. Concrete Implementation Steps

### Step 10.1: Production Dockerfile Generator
1. Implement multi-stage build:
   - **Builder Stage**: `golang:1.23-alpine` with build cache mounts (`--mount=type=cache,target=/go/pkg/mod`).
   - Compiles static binary: `CGO_ENABLED=0 go build -ldflags="-s -w" -o /bin/app ./cmd/api`.
   - **Runtime Stage**: Distroless or `alpine:3.20` with non-root user (`USER 10001:10001`), ca-certificates, and tzdata.
2. Provide local `docker-compose.yml`:
   - Spin up PostgreSQL and Valkey with healthchecks and persistent volumes.

### Step 10.2: Kubernetes Manifest Generator
1. Generate standard production Kubernetes resources in `deploy/k8s/`:
   - `Deployment`: Enforces memory/CPU resource requests & limits; sets non-root security context; configures `livenessProbe` (`/health/live`) and `readinessProbe` (`/health/ready`).
   - `Service`: Standard ClusterIP exposing HTTP port 80/8080.
   - `Ingress`: Configurable TLS and host routing rules.

### Step 10.3: CI Pipeline Template Generator
1. Scaffold `.github/workflows/ci.yml`:
   - Run `go vet ./...`.
   - Run `loy check`.
   - Run `go test -v -race ./...`.
   - Build binary and run container smoke test.

### Step 10.4: Canonical Release Gate Verification
Execute the full canonical verification loop:
```bash
loy new demoapp --preset api
cd demoapp
loy make crud products
loy check
go test ./...
go build ./...
docker build -t demoapp:latest .
```

---

## 4. Test Strategy & Acceptance Criteria

### Container Tests
- Build Docker image -> inspect image size (< 30MB compressed).
- Run container -> verify binary starts under non-root UID.
- Send SIGTERM to container -> verify clean shutdown within 5 seconds.

### Kubernetes Manifest Validation
- Run `kubeval` or `kubectl --dry-run=client -f deploy/k8s/` -> verify zero syntax or schema errors.

---

## 5. Definition of Done
- [ ] Multi-stage Dockerfile builds and runs cleanly.
- [ ] Kubernetes manifests deploy cleanly with valid liveness and readiness probes.
- [ ] End-to-end acceptance loop passes completely from scratch.

---

[← Previous: Phase 9 Plan](./09-Phase-Fullstack.md) | [Back to Plans Index](./README.md)
