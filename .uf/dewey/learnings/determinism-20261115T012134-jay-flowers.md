---
tag: determinism
author: jay-flowers
category: pattern
created_at: 2026-11-15T01:21:34Z
identity: determinism-20261115T012134-jay-flowers
tier: draft
---

When a human-confirmation gate mis-branches, fixing it by adding more classification prose compounds the drift risk that caused the incident. Observed in /uf.review-pr (issue #1055, fix PR #1077): the same-verdict duplicate gate recommended skipping a re-review on a new head commit because the prose gate treated any same-verdict prior review as a duplicate; the fix rewrote Step 8a with commit_id == head_sha comparisons, finding-set diffs, and branch rules as ~45 lines of prose (+632 tokens), even though the command's own Step 3.11 already delegates prior-review-vs-head analysis to the tested `resolve_review_mode` tool ('The tool owns the decision; do not restate, recompute, or override it'). Twice-test for gate branches: when a gate's branch selection is a pure function of structured data (SHA equality, finding-set diff), extract the classification to a tool that returns the branch plus the diff details, and let the gate prose consume it — only the presentation, option reasoning, and recommendation wording are judgment that stays in prose. Prose-executed classification is also the layer where auto-question accepts weak-confidence answers (0.44 in the incident) because the gate's recommendation itself was wrong.
