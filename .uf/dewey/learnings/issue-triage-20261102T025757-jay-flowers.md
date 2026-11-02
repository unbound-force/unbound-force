---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T02:57:57Z
identity: issue-triage-20261102T025757-jay-flowers
tier: draft
---

When triaging determinism/tool-extraction issues, verify whether the proposed deterministic logic already exists in an adjacent tool before accepting a re-implementation plan. In unbound-force, the review-dispatch plugin already computes path→category change profiles (7 categories plus security_sensitive, user_facing, and tier derivation) for plan_review_dispatch/dispatch_review_runs; issue #673's classify_paths tool is therefore valid only as an extraction of that classifier to .opencode/lib/ as a shared pure function (classifyBranch/matchTaskLine/sanitizeSlug precedent) with at most a thin wrapper — re-implementing the patterns in uf-workflow would create two classifiers over the same taxonomy, and when they drift, dispatch plan tiers (reviewer selection, model tiering) and command-level review focus silently disagree.
