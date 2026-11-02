---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T04:36:50Z
identity: issue-triage-20261102T043650-jay-flowers
tier: draft
---

Triage pattern for review-pipeline gap claims: an existing "intent drift" reviewer mandate (e.g., Divisor Guard: "detect intent drift against active stories and acceptance criteria") is spec-keyed — it validates the PR against spec artifacts resolved from the branch, not against the originating GitHub issue. A PR can satisfy every quality gate and pass spec-alignment review while drifting from the originating issue's acceptance criteria, because (a) issue linking during review-context gathering is optional and PR-number-conditional, and (b) the no-spec fallback uses the PR title and description as the intent source, never the issue. When triaging a proposed intent-satisfaction check, verify three things before accepting it as actionable: machine-readable issue-ID threading from pipeline start to review time (the check is meaningless without it), grounding evidence whose rigor level is itself validated (intake-kit-tested Given/When/Then ACs vs. best-effort semantic evaluation, with the degradation documented), and a governance escape hatch (typed deviation routed to amend the requirement) for legitimately wrong acceptance criteria rather than a binary pass/fail verdict.
