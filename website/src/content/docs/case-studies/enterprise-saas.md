---
title: "Case Study: Building Intivai Enterprise SaaS with Loy"
description: "How an enterprise AI hiring platform scales across 12 bounded contexts using Loy."
---

**Intivai** is an enterprise AI talent screening platform handling resume parsing, distributed candidate screening, WebSocket interview proctoring, and automated rubric scoring.

This case study illustrates how Intivai leverages Loy's architecture patterns to maintain high velocity and zero technical debt.

## System Topography

```
Intivai Production Fleet
├── cmd/api/ (Public Fiber REST API + WebSockets on :8080)
├── cmd/worker/ (Asynchronous Asynq Consumer Daemon)
└── 12 Modular Bounded Contexts:
    ├── IAM & Organizations (internal/iam/)
    ├── Candidate Pipeline (internal/candidate/)
    ├── Job Postings (internal/job/)
    ├── Resume Parsing (internal/cv_parser/)
    ├── Real-Time Proctoring (internal/proctoring/)
    └── Rubric Evaluation (internal/evaluation/)
```

---

## 1. Multi-Tenant Isolation with PostgreSQL RLS

Every interview and evaluation is secured with PostgreSQL Row Level Security ([ADR-018](/loy/adrs/)).

- Tenant identity is resolved from JWT claims.
- Repositories set `app.current_tenant_id` inside transactions.
- Zero risk of cross-organization candidate data leaks.

---

## 2. Modular Sub-Domain Composition Roots

With over 35 distinct services and repositories, monolithic `wiring.go` files become unmaintainable. Intivai uses modular wire files ([ADR-020](/loy/adrs/)):

```
internal/app/
├── wiring.go            # Coordinates sub-domain wire calls (< 40 LOC)
├── wire_iam.go          # IAM domain composition root
├── wire_candidate.go    # Candidate domain composition root
├── wire_job.go          # Job posting composition root
└── wire_proctoring.go   # Proctoring WebSocket composition root
```

---

## 3. Real-Time WebSocket Interview Proctoring

Intivai uses Loy's WebSocket scaffolding (`loy make ws proctoring`) to stream tab-focus telemetry, microphone status, and AI evaluation tokens with dedicated read/write pumps and heartbeat monitors.

---

## 4. Architecture Enforcement in CI

Every pull request runs `loy check --strict` in GitHub Actions, preventing layer direction violations or cross-domain package coupling before code is merged.
