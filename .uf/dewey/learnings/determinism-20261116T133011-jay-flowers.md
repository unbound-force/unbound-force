---
tag: determinism
author: jay-flowers
category: pattern
created_at: 2026-11-16T13:30:11Z
identity: determinism-20261116T133011-jay-flowers
tier: draft
as_of: "unbound-force@db9c0c04b4e23c28da65ae3f5c5209e85114182a"
---

Cross-plugin error-classification drift is cured by one shared decision table, not by cross-pinning two private copies. In unbound-force issue #1023, gh exit code 4 classified as rate_limited/retryable in review-dispatch but gh_auth_failed/not-retryable in feedback-replies — one failure class, opposite retry decisions. The durable fix (openspec/changes/fix-feedback-reply-followups): a pure-function shared lib module (classifyGhFailure in .opencode/lib/gh-error.ts) holding the single decision table — stderr text keys first, the documented exit code as fallback — consumed through thin adapters that keep their own message templates and annotate retryability onto the error so consumers never re-derive it, pinned by a decision-table test plus adapter-parity tests running identical inputs through each consumer's public surface, and registered in the scaffold (mirror, isToolOwned/isDivisorAsset, stageReviewPluginSources) in the same increment so scaffolded repos stay byte-compatible.
