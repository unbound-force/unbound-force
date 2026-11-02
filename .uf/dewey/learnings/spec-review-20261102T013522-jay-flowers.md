---
tag: spec-review
author: jay-flowers
category: pattern
created_at: 2026-11-02T01:35:22Z
identity: spec-review-20261102T013522-jay-flowers
tier: draft
---

When a design decision is strengthened mid-flight (e.g., expanding a CLI flag-conflict set from three to four flags), the operative artifacts (spec delta FRs, tasks) usually get updated, but the proposal's restatements — What Changes items and the Done-means acceptance summary — are easy to miss because they are prose context rather than normative text. In OpenSpec changes, after any design-level strengthening, grep every artifact in the change directory for each restatement of the acceptance contract (conflict sets, preservation guarantees, file lists); the proposal is what reviewers verify against. Caught in the init-override-stubs-gitignore-sync spec review: proposal.md omitted --stealth from the --overrides conflict set (lines 70 and 106-107) while design D5, SC-FR-012, and task 4.2 all carried four conflicts.
