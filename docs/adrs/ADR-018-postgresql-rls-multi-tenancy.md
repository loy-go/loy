# ADR-018: PostgreSQL RLS Multi-Tenancy Engine

## Status
Accepted

## Context
Multi-tenant enterprise SaaS applications require strict data isolation between organizations. Manually adding `WHERE org_id = $1` to every SQL query is error-prone, pollutes query signatures, and risks cross-tenant data leaks.

## Decision
1. Provide first-class multi-tenancy configuration in `loy.yaml` (`multi_tenancy: { enabled: true, strategy: rls }`).
2. Code generators inject `org_id UUID NOT NULL` and `FORCE ROW LEVEL SECURITY` policies directly into PostgreSQL migrations.
3. Database isolation is enforced transparently in PostgreSQL via `SET LOCAL app.current_tenant_id` session hooks within transactions.
4. Keep sqlc queries and domain repository interfaces clean of synthetic tenant query arguments.

## Consequences
- Transparent database-level tenant isolation with zero chance of query filtering omissions.
- Queries and domain signatures remain clean and portable.
- Zero runtime framework dependencies introduced.

---

[Back to ADR Index](./README.md) | [Back to Documentation Index](../00-INDEX.md)
