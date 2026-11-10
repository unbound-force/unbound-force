---
tag: instruction-quality
author: jay-flowers
category: pattern
created_at: 2026-11-08T23:16:49Z
identity: instruction-quality-20261108T231649-jay-flowers
tier: draft
---

When an instruction file already contains one mechanism that maps a value (verdict to event) and a change adds or extends a second mechanism producing the same output (e.g., a self-authored remap that now also captures severity-mapped advisory approvals), every condition phrased on the mapping's result — "when the verdict maps to COMMENT", "APPROVE verdicts only" — becomes ambiguous or stale. Authoring discipline: when adding a new mapping path, audit every consumer of the old invariant (warning triggers, option-set selectors, body-note conditions) in the same edit, and phrase conditions on the producing mechanism or the underlying condition ("when any finding is CRITICAL or HIGH"), never on the mapping result alone. Observed on opsx/advisory-approval-event (PR #990): the new severity-keyed mapping in uf.review-pr.md Step 8c left "When the verdict maps to COMMENT" resolvable against both the severity table and the self-authored remap, and left Step 8a's "(APPROVE verdicts only)" checks not covering the new class of posted approvals.
