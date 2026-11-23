---
tag: determinism
author: jay-flowers
category: pattern
created_at: 2026-11-16T14:28:11Z
identity: determinism-20261116T142811-jay-flowers
tier: draft
as_of: "unbound-force@2318d58fcc5416396271851f3e9bf2ef7c021faf"
---

When duplicated decision-table logic across plugin families is consolidated into one shared lib module (fix-feedback-reply-followups: classifyGhFailure in .opencode/lib/gh-error.ts, replacing the divergent exit-4 mappings in feedback-replies and review-dispatch), no registered tool surface changes, so `tool-inventory --seams` reports an empty list and DE-007's registration/output-shape/round-trip layers do not apply — the change is code-to-code, not a prose→tool migration. The seam of record for such a consolidation is the adapter-parity pin: decision-table rows for the shared pure function plus tests that run identical signal inputs through each consumer's public surface (classifyGhError, the fetch_pr_checks tool) and assert identical code and retryability — the shape .opencode/test/gh-error.test.ts established.
