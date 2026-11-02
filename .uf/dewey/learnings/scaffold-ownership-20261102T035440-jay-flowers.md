---
tag: scaffold-ownership
author: jay-flowers
category: gotcha
created_at: 2026-11-02T03:54:40Z
identity: scaffold-ownership-20261102T035440-jay-flowers
tier: draft
---

unbound-force #649 triage (uf/review-matrix.yaml stale-matrix skip): a user-owned scaffold classification is an intentional extension-point decision, but the defect lives in what the skip path omits — it appends to result.Skipped without comparing the canonical embedded asset to the on-disk file, so schema-version skew (v2 flat provider strings vs v3 structured model/advisor/runs/risk_augmentation/limits sections) is never surfaced and downstream review dispatch (review-council, triage-issue, address-feedback) breaks silently. Verified in internal/scaffold/scaffold.go: --force overwrites every non-isNeverOverwrite asset, so the review-matrix branch comment "without --force replacing their policy" states a protection that does not exist — only isNeverOverwrite files (constitution.md) survive --force. Sibling branches added later (uf/model-matrix.yaml, issue #889) carry the corrected phrasing ("--force remains the explicit replace path"), showing comment-accuracy drift accumulates when older branches are not updated alongside newer ones. Rule: a user-owned skip branch for a versioned-schema asset must compare content and warn on divergence (or pair the skip with a fail-closed downstream error naming the file to edit, per the #884 remediation pattern), and any comment describing --force behavior must match the force-mode code path.
