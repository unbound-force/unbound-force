---
tag: feedback-triage
author: jay-flowers
category: gotcha
created_at: 2026-11-07T20:50:47Z
identity: feedback-triage-20261107T205047-jay-flowers
tier: draft
---

Triaging record-semantics feedback on PR 948 (plugin-package-rollback.json baseline_lock_sha256): when a review-fix round deliberately redefines a governance record field's semantics mid-PR (rollback-target lock SHA -> current-state lock SHA, so the revert baseline stays executable), the new semantics get documented at the enforcement site (test comment + executable hash anchor) while the definition site — the unarchived spec MUST-text — keeps the superseded semantics and a stale hash, leaving an active spec contradicted by the shipped tree. Two triage disciplines: (1) check cross-item convergence before proposing an independent fix — PR 948 feedback items 2 and 8 both targeted the SC-FR-003 rollback clause, so a single superseded-by annotation resolves both; (2) scope "recorded nowhere in-tree" absence claims by diffing removed values against retained record fields — the claim was true for the pre-bump hash (1ac9cdf1…, removed in lockstep from testdata and test const; rollback only via git revert of the bump commit) but false for the pre-hardening hash (fef6f7f1…, retained in security_override_rollback.lock_sha256 and its test constant).
