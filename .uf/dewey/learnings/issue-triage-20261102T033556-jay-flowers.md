---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T03:35:56Z
identity: issue-triage-20261102T033556-jay-flowers
tier: draft
---

Duplicate-candidate scans over open issues cannot surface resolutions that shipped as code rather than as issues — an issue can be overtaken by events while every issue-level keyword search returns only itself. unbound-force issue #658 (2026-10-05, proposing a purpose-built review-orchestrator agent to replace cobalt-crush-dev in /uf.unleash review-council delegation) exemplifies this: the repo instead eliminated the intermediary entirely — the review steps now dispatch directly via plan_review_dispatch → dispatch_review_runs (PR #729, opsx/unleash-direct-review-dispatch, documented in learning uf-unleash-20261006T004544) — so the issue's cited delegation sites no longer exist in the current command, yet the duplicate-candidate scan reported "none found". Triage procedure that catches obsolescence: (1) read the issue's cited file:line sites in the current working tree — content mismatch or line drift signals restructuring since filing; (2) query Dewey for the issue's topic keywords — prior learnings and merged PRs reveal shipped alternatives that issue-only searches miss; (3) when the premise sites have been rewritten, return INVALID as overtaken-by-events rather than treating the issue as actionable as written.
