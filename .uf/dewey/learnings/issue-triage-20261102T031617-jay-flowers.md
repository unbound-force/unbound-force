---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T03:16:17Z
identity: issue-triage-20261102T031617-jay-flowers
tier: draft
---

Triage verification of unbound-force tool-extraction issues must grep for procedure signatures, not citation tokens — and underspecified mechanics strengthen the extraction case. Issue #669 (ensure_labels, fetch_issue_comments) cited `gh label list` and `gh api --paginate --slurp`; current HEAD contains neither token, yet uf.triage-issue.md Section 4.2 verifiably prescribes the dual-path label procedure (Path A `gh label create` + `gh issue edit --add-label`, Path B apply-existing) with NO explicit repository-label existence command — the existence check is left to LLM improvisation, a stronger DE-001 violation than prose duplication alone. The comment-pagination prose survives in both cited files (uf.triage-issue.md ~L68-94 with per-page validation choreography; uf.address-feedback.md ~L146-160, four --paginate calls). Naming drift also appears in test criteria: "mock BunShell" names a primitive the repo deliberately avoids (review-dispatch L7605 comment: Bun.spawn, NOT BunShell template literals); the real pattern is createXxxTool(exec: GitExec, cwd) factories with injected exec mocked in tests, and uf-workflow already carries a gh-execution signature mirroring GitExec (~L1318, the uf_unleash_resolve_issue precedent). Token-level citation drift is a plan-time refinement, not a NEEDS-CLARIFICATION blocker, when the procedure signature and acceptance-criteria intent verifiably exist.
