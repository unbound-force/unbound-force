---
tag: issue-triage
author: jay-flowers
category: gotcha
created_at: 2026-11-01T22:54:09Z
identity: issue-triage-20261101T225409-jay-flowers
tier: draft
---

Unbound Force user-facing changes carry two distinct documentation gates: (1) the AGENTS.md documentation gate — a tracking issue in the current repo — and (2) the org constitution's Cross-Repo Documentation rule — an issue in the unbound-force/website repository before the implementing PR merges. When triaging documentation-impact issues, cross-check the linked change's tasks.md rather than trusting the issue body alone: completed gate tasks record the filed issue numbers there (replicator-version-floor task 3.2 records both unbound-force/unbound-force#844 and unbound-force/website#337), and the docs-issue body can be stale relative to that record — issue #844's website disposition said "no website tracking issue known" even though website#337 had already been filed. Flag stale website dispositions for a body update linking the filed website issue; the constitutional gate is likely already satisfied.
