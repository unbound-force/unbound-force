---
tag: per-task-delegation-budgets
author: jay-flowers
category: review-insight
created_at: 2026-11-16T09:55:00Z
identity: per-task-delegation-budgets-20261116T095500-jay-flowers
tier: draft
---

The opsx/per-task-delegation-budgets code review (correlations bd99036f then ea364acd, 2026-11-16) showed the fastest path through a REQUEST CHANGES verdict is to relocate mechanism into the tool layer rather than defend instruction prose. Iteration 1 returned seven findings; the CRITICAL one (DE-003) flagged a phase-checkpoint budget line the command prose told the agent to reconstruct from a three-slot template, and a MEDIUM one (DE-006) flagged retry routing re-derived from timeout.kind enumerations. Both were resolved in a single fix delegation by (a) exporting a pure renderPhaseCheckpointLine from the uf-workflow plugin so the checkpoint tool itself renders `Phase [N] checkpoint: [verdict] (remaining budget: ~[X]ms)` and the command emits the returned phase_checkpoint_line verbatim, and (b) rewriting the retry branch to consume the envelope's already-computed error.retryable field. The scaffold content gate was updated to pin the corrected contracts and strengthened with stale-absence checks (the old template and kind-enumeration prose now fail the gate if they return) — fixing a determinism finding should strengthen the pin, never weaken it. Iteration 2 returned APPROVE WITH ADVISORIES with seven advisory findings (6 MEDIUM, 1 LOW); the human gate accepted them as carry-forward rather than fix-forward, which respects the no-code-changes-between-APPROVE-and-PR rule, since advisory fixes after an approving verdict would require a fresh re-review to keep the reviewed artifact identical to the merged one. Tags: opsx/per-task-delegation-budgets, 2026-11-16, review-insight.
