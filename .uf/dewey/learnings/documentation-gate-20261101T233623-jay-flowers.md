---
tag: documentation-gate
author: jay-flowers
category: pattern
created_at: 2026-11-01T23:36:23Z
identity: documentation-gate-20261101T233623-jay-flowers
tier: draft
---

A per-change documentation-gate issue is well-formed when it: (1) records per-file dispositions with reasoning — residual decisions tracked explicitly, pre-existing gaps deferred to their existing trackers rather than duplicated; (2) performs an explicit pre-filing overlap check against related gate issues (parent-change gate issue, epic trackers, website-side coverage); and (3) reuses existing cross-repo website coverage instead of filing a duplicate when the change extends an existing contract rather than adding a documentation surface — recording any staleness in the reused issue so eventual documentation reflects the extended contract. Validated triaging unbound-force/unbound-force#834 (gate record for spawn-subtask-classification-params): every material claim — CHANGELOG completion in-change, per-file dispositions, overlap check, website-coverage reuse via unbound-force/website#317 — was corroborated by the change's proposal.md Task 4.2 record in the Dewey-indexed knowledge graph, and the unexplained #787 refs-line entry resolved to the parent delegation-skill-enforcement change via the change's design context.
