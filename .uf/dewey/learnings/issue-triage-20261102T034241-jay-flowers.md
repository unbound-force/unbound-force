---
tag: issue-triage
author: jay-flowers
category: gotcha
created_at: 2026-11-02T03:42:41Z
identity: issue-triage-20261102T034241-jay-flowers
tier: draft
---

Divisor agent inventory drift in unbound-force: when triaging agent-refactor issues, verify the persona roster against the filesystem and .uf/review-matrix.yaml rather than trusting the issue text. Three counts coexist — issue #660 named 5 Divisor agents (a framing inherited from spec 019's era, when the council was five personas), the review-matrix advisor roster for every mode is 6 (includes divisor-architect), and 9 divisor-* files under .opencode/agents/ carry the shared boilerplate (divisor-envoy, divisor-herald, divisor-scribe were added later). Per-section duplication also varies: "Step 0: Prior Learnings" appears in all 9 files, "## Decision Criteria" in 6 divisor files, "### Out of Scope" in exactly 5 (adversary, architect, guard, sre, testing — not curator). An acceptance criterion scoped to a stale roster leaves the later-added personas carrying inline duplicates, defeating the DRY goal with a two-tier inconsistency.
