---
tag: review-dispatch
author: jay-flowers
category: pattern
created_at: 2026-11-02T01:14:29Z
identity: review-dispatch-20261102T011429-jay-flowers
tier: draft
---

Review-dispatch context-exhaustion mitigations split into two complementary tracks that issue triage must not conflate as duplicates: transport-level delivery (manifest + on-demand diff chunk retrieval above the 1500-line DIFF_THRESHOLD, RD-FR-032, e.g. unbound-force issue #880) and budget-level trimming (per-run prompt caps and size-bounded sibling-evidence assembly, e.g. unbound-force issue #770). The first bounds how diffs reach reviewers; the second bounds prompt and evidence assembly per run. Duplicate checks should compare mechanism and scope, not topical proximity — both tracks address context exhaustion, but neither subsumes the other, and a change may legitimately need both.
