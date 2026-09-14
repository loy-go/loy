---
title: "Production Readiness & Hardening Guide"
description: "Zero-downtime migrations, graceful shutdown budgets, and observability playbooks."
---

Deploying mission-critical Go systems requires strict operational hygiene. This guide covers production playbooks built into Loy runtime scaffolding.

## 1. Graceful Shutdown & Kubernetes Drain Budgets

Loy runtime scaffolds a phased LIFO `shutdown.Coordinator` in `internal/platform/shutdown/coordinator.go` ([ADR-004](/loy/adrs/)):

```
SIGTERM Received ──► Stop Accepting New Traffic (HTTP/WS Listener Closed)
                 ──► Drain In-Flight Requests (Grace Period: 15s)
                 ──► Flush Telemetry & Traces (OpenTelemetry Provider)
                 ──► Close Database Connection Pools & Queue Consumers
                 ──► Process Exit (0)
```

In your Kubernetes Deployment, ensure `terminationGracePeriodSeconds` exceeds your application `SHUTDOWN_TIMEOUT`:

```yaml title="k8s/deployment.yaml"
spec:
  template:
    spec:
      terminationGracePeriodSeconds: 30
      containers:
        - name: api
          env:
            - name: SHUTDOWN_TIMEOUT
              value: "20s"
```

---

## 2. Zero-Downtime Migration Strategy

When deploying schema migrations in production with `loy migrate up`:

1. **Add Columns as Nullable or with Defaults:** Never add non-null columns without defaults to high-traffic tables.
2. **Expand and Contract:** Deploy code that writes to both old and new columns before removing deprecated columns.
3. **Run Migrations via Pre-Deploy Hooks:** In Kubernetes, execute migrations inside a `Job` or Helm `pre-upgrade` hook before rolling new pods.

---

## 3. Database Connection Pooling Parameters

Configure pgx connection pool parameters via environment variables in `internal/config/config.go`:

| Variable | Recommended Production Value | Description |
|---|---|---|
| `DB_MAX_CONNS` | `25` - `50` per replica | Maximum concurrent connections per pod |
| `DB_MIN_CONNS` | `5` | Minimum idle connections kept warm |
| `DB_MAX_CONN_LIFETIME` | `1h` | Prevents stale connections and load balancer timeouts |
| `DB_MAX_CONN_IDLE_TIME` | `30m` | Reclaims idle connection resources |
