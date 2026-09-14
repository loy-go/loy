# Loy Project Governance

This document outlines the governance model, roles, decision-making process, and RFC mechanism for the Loy open-source project.

---

## 1. Governance Model: Maintainer-Led with RFCs

Loy operates under a **Maintainer-Led** governance model. Strategic direction and day-to-day project health are stewarded by designated Maintainers with a Project Lead coordinating releases and consensus. Major architectural shifts, breaking CLI changes, and new generator paradigms require an RFC (Request for Comments) process with public review.

---

## 2. Roles & Responsibilities

### Users
Anyone who downloads, runs, or uses Loy in personal or commercial projects. Users contribute by filing bug reports, requesting features, and participating in discussions.

### Contributors
Anyone who submits a pull request, updates documentation, improves test coverage, or reviews open issues. Contributors must follow [CONTRIBUTING.md](CONTRIBUTING.md) and the [Code of Conduct](CODE_OF_CONDUCT.md).

### Maintainers
Active contributors with write access to the repository. Responsibilities include:
- Reviewing and merging pull requests.
- Triaging issues and security reports.
- Ensuring compliance with architectural invariants (ADRs).
- Mentoring new contributors.

**Becoming a Maintainer:**
Active contributors who demonstrate sustained technical excellence, adherence to project conventions, constructive code reviews, and community empathy over at least 3 months may be nominated by an existing Maintainer. Approval requires a supermajority (2/3) vote of current Maintainers.

### Project Lead
The Project Lead coordinates release schedules, resolves unresolvable ties in technical debates, manages organizational credentials, and acts as the point of contact for external partnerships and security escalations.

---

## 3. Decision-Making & Consensus

- **Routine Changes**: Bug fixes, minor documentation updates, non-breaking performance optimizations, and rule tweaks require approval from at least one Maintainer.
- **Architectural & Breaking Changes**: Any modification to core invariants, public CLI flags, JSON output schemas, generator contracts, or rule semantics requires an approved RFC.

---

## 4. RFC (Request for Comments) Process

When proposing significant changes:
1. **Open an Issue**: Tag it with `kind/rfc` and outline the problem statement, motivations, and non-goals.
2. **Draft the Proposal**: Document proposed design following the ADR template in `docs/adrs/`.
3. **Public Review**: The RFC remains open for community review and feedback for at least 7 days.
4. **Resolution**: Maintainers vote to accept, request revisions, or decline the RFC. Accepted RFCs are assigned an official ADR number and merged into `docs/adrs/`.

---

## 5. Conflict Resolution

If consensus cannot be reached after thorough discussion:
1. Maintainers seek guidance from the original ADR authors.
2. If disagreement persists, the Project Lead casts the deciding vote to break deadlocks and maintain velocity.
