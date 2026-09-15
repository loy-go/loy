---
title: "Prometheus Metrics & Grafana Dashboards"
description: "Turnkey RED metrics recording, Prometheus scrapers, and production Grafana dashboards via loy make metrics."
---

Observability shouldn't be an afterthought added during an incident. Every production Go service must expose the standard **RED Method** metrics:
- **Rate**: number of HTTP requests handled per second.
- **Errors**: number of requests that fail with 4xx or 5xx status codes.
- **Duration**: latency distributions (P50, P90, P95, P99).

Loy provides an all-in-one generator:
```bash
loy make metrics [name]
```

---

## Generated Observability Stack

Running `loy make metrics` scaffolds three synchronized components:

```bash
loy make metrics
```

```text
Successfully generated metrics app
  + internal/platform/metrics/metrics.go (create)
  + deploy/grafana/dashboard.json (create)
  + deploy/prometheus.yml (create)
```

### 1. Prometheus Platform Recorder (`internal/platform/metrics/metrics.go`)

A pure Go metrics recorder implementing standard Prometheus collectors:

- `http_requests_total`: Counter partitioned by `method`, `path`, and `status`.
- `http_request_duration_seconds`: Histogram with exponential buckets tailored for sub-millisecond to multi-second latency tracking.
- `active_connections`: Gauge tracking concurrent inflight requests.
- `PrometheusMiddleware`: Drop-in HTTP middleware compatible with Fiber, Chi, Gin, and net/http.
- `PrometheusHandler`: Exposes `/metrics` endpoint for Prometheus scrapers.

### 2. Turnkey Grafana Dashboard (`deploy/grafana/dashboard.json`)

A production-ready JSON dashboard definition featuring:
- **Throughput Panel**: Real-time QPS split by HTTP method and status code.
- **Error Rate Gauge**: Percentage of 5xx server errors over 5m rolling windows.
- **Latency Percentiles**: P50, P95, and P99 latency graphs over time.
- **Go Runtime Telemetry**: Active goroutines, heap allocations, and GC pause durations.

Simply import `deploy/grafana/dashboard.json` into Grafana to get instant visualization.

### 3. Prometheus Scrape Configuration (`deploy/prometheus.yml`)

A ready-to-run scraper config targeting your local or Kubernetes deployment on port 8080.

---

## Wiring into the Runtime

To attach the metrics recorder to your HTTP server, update `internal/app/wiring.go`:

```go
import "github.com/example/app/internal/platform/metrics"

// Inside WireApp:
metricsRecorder := metrics.NewRecorder()
app.Use(metricsRecorder.Middleware())
app.Get("/metrics", metricsRecorder.Handler())
```

---

## Correlating Traces with OpenTelemetry

When combined with OpenTelemetry (`loy make runtime`), Loy automatically attaches W3C `traceparent` context to outgoing responses and metric attributes, allowing you to jump from a P99 latency spike in Grafana directly into Jaeger or Tempo traces.
