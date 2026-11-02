---
tag: review-dispatch
author: jay-flowers
category: pattern
created_at: 2026-11-02T03:51:41Z
identity: review-dispatch-20261102T035141-jay-flowers
tier: draft
---

Mechanically-verified byte-copies carry zero marginal review signal. In unbound-force, files under internal/scaffold/assets/ are byte-copies of canonical .opencode/ sources, and TestEmbeddedAssets_MatchSource enforces byte-identical parity as a CI gate (a MUST per openspec/changes/fix-ci-embedded-drift). A diff that only syncs these copies is fully verified by the drift gate plus canonical-source review, yet review-dispatch classifies them "standard" and counts their lines toward tier — routine syncs then inflate the review tier (PR #653: 1083 of 1337 diff lines were scaffold syncs, pushing a ~20-meaningful-line change to heavy). When a deterministic test already proves file content, agent re-review of that content is redundant (Constitution Principle VI); classify such mirrors low-signal, and make the change in the single canonical classifier (per the single-classifier principle, see issue #673's extraction direction) rather than a parallel rule.
