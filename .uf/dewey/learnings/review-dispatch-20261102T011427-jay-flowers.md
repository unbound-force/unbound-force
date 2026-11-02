---
tag: review-dispatch
author: jay-flowers
category: gotcha
created_at: 2026-11-02T01:14:27Z
identity: review-dispatch-20261102T011427-jay-flowers
tier: draft
---

Review-dispatch budget enforcement must ship with its audit trail in the same increment. The fix-review-dispatch-recovery change (issue #751) deliberately sequenced observability first — RD-FR-027 surfaces per-run token usage in dispatch_status — and deferred prompt/evidence trimming to issue #770. When extending budget_usd usage-aware skipping to per-run token/context budgets: (1) budget-skipped reviewers must surface as budget_skipped run status in consolidation, never silently omitted, or the budget silently disables review coverage — a security-relevant bypass control; (2) sibling-evidence truncation must preserve the untrusted-evidence delimiters and provenance (sibling, commit, sha256) so trimmed evidence stays verifiable; (3) existing evidence bounds are per-file, not per-run aggregate — the aggregate (files × reviewers) is the unbounded axis that per-run budgets must cap.
