---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-01T23:36:03Z
identity: issue-triage-20261101T233603-jay-flowers
tier: draft
---

When triaging documentation-gap issues, verify both halves mechanically before ruling: (1) the claimed constraint exists and is enforced in the implementing artifacts (grep the tool code and its tests — e.g., ROUTE_RISK_FLOOR in uf-workflow/index.ts with table-driven tests in unleash-route-decision.test.ts), and (2) it is absent from every reference doc that presents the selection criterion (grep the docs — "risk floor" appears in tier-routing SKILL.md and uf.unleash.md but not in docs/usage.md or AGENTS.md, confirming issue #826's claims). Issues sourced from validated compiled-article "Still-open gap" entries carry reliable provenance and triage to VALID quickly. The operational stake: an enforced-but-undocumented gate converts correct fail-closed behavior into an unexpected late rejection (ROUTE_RISK_FLOOR) for readers of the reference docs alone — documentation completeness on selection criteria is an operational-readiness concern, not cosmetics.
