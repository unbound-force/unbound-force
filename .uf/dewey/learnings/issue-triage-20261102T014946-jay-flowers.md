---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T01:49:46Z
identity: issue-triage-20261102T014946-jay-flowers
tier: draft
---

Bug-report triage in fast-moving repositories must verify cited evidence at both the pinned commit the report references and current HEAD, because "accurate when filed" and "still unfixed" are independent questions. unbound-force issue #746 (review-dispatch's one-dispatch-session-per-conversation silently breaking uf.address-feedback per-item Tier 2 finalization) was accurate at head f0327a9 — command-contract quotes verbatim, first-call-only session_metadata persistence corroborated by a pre-filing Dewey learning and sibling issue #760 — yet both of its suggested fix directions landed within two weeks (design D5 replace-with-latest metadata in fix-review-dispatch-recovery; design D1 per-invocation correlation-ID minting in fix-review-dispatch-correlation-release, shared root cause #794), leaving it a close-as-completed candidate rather than an open defect. Classify such reports VALID with an already-remediated status note instead of INVALID: rejecting an accurate report erases the audit trail, while re-fixing a remedied defect wastes dispatch budget. Fastest remediation provenance: Dewey learnings plus active openspec/changes/ directories on disk.
