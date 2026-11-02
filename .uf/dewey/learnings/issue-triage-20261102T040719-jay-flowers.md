---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T04:07:19Z
identity: issue-triage-20261102T040719-jay-flowers
tier: draft
---

Triage feature issues against current repository state, not only against duplicate candidates. Issue #636 (unbound-force meta repo, filed 2026-09-30) requested a declarative sibling-repository contract plus a downstream-impact review persona; by triage time the repository had already shipped the core mechanism — .uf/sibling-repos.yaml, the acquire_sibling_evidence tool (origin-verified, commit-pinned, untrusted-evidence delimited), and sibling_evidence injection into review dispatch — but in a different architectural shape (evidence distributed to every dispatched reviewer rather than consumed by a dedicated persona). Duplicate-only triage misses issues overtaken by implementation and misroutes delivered work as new work; the correct disposition is close-as-completed or re-scope to the remaining delta, decided by comparing the issue's proposed design against the shipped design.
