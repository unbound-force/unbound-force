---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-01T23:36:01Z
identity: issue-triage-20261101T233601-jay-flowers
tier: draft
---

When triaging documentation-gap issues about governance constraints, verify both the gap and the enforcement path before assessing impact. In unbound-force issue #826 (tier-routing risk floors missing from the docs/usage.md workflow table and the AGENTS.md Quick row), the floors are mechanically enforced in uf_unleash_route_decision (ROUTE_RISK_FLOOR, pinned by table-driven tests in unleash-route-decision.test.ts), so /uf.unleash-mediated routing fails closed — the doc gap cannot cause a mis-route through the tool itself. But the selection docs remain load-bearing for the manual path: an engineer reading only the Quick row can hand-create a quick/<name> branch, and uf_unleash_branch_gate accepts quick branches with no tier prerequisite and no floor re-check (floors live only in the route-decision step). Pattern: mechanical enforcement bounds a doc gap's tool-mediated impact, yet selection docs stay accuracy-critical wherever a human can act on them without passing through the enforcing tool. A summary-or-pointer fix in every doc presenting the selection criterion remains proportionate and necessary.
