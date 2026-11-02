---
tag: issue-triage
author: jay-flowers
category: gotcha
created_at: 2026-11-02T03:54:52Z
identity: issue-triage-20261102T035452-jay-flowers
tier: draft
---

Feature-request issues that ask to port a capability from another checkout can arrive stale in fast-moving repositories: the port may land and even be refined before triage runs. Before returning VALID on such an issue, verify the premise against HEAD — read the target command/agent file, check convention packs for present-tense implementation claims (e.g., severity.md's "This policy is implemented in /uf.review-council Spec Review Mode"), and check for completed openspec/changes/<name>/tasks.md artifacts (all checkboxes [x]) from later issues refining the same capability. Case: unbound-force issue #637 (spec review mode for the review council, filed 2026-09-30) was fully implemented at triage time 2026-11-01 — uf.review-council.md:1175-1400 carries the complete Spec Review Mode (tier-dependent scope, auto-detection, specs-mode dispatch, hybrid fix policy, 3-iteration cap), and the spec-review-per-finding-gate change (originating issue #777) had already refined its gate into a per-finding protocol.
