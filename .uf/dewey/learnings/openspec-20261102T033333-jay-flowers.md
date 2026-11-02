---
tag: openspec
author: jay-flowers
category: reference
created_at: 2026-11-02T03:33:33Z
identity: openspec-20261102T033333-jay-flowers
tier: draft
---

Requirement-numbering allocation notes (meta-prompt pack MP-014) appear in this repo only on deltas that allocate FR ids: derive-per-run-timeout/specs/review-matrix.md (RD-FR-009/010), optional-profile-models/specs/review-matrix.md (RD-FR-019..024, as a `> **Numbering**:` blockquote), the model-matrix-delegation deltas, and host-fallback-provenance's review-dispatch delta (RD-FR-037). A change may place the note in one delta covering the whole change (derive-per-run-timeout puts it in review-matrix.md, not its review-dispatch delta), and id-allocating changes are not uniformly compliant (review-dispatch-diff-manifest and fix-review-dispatch-recovery carry no notes). To verify an RD-FR allocation is collision-free, grep across openspec/changes/*/specs in both flat and directory layouts.
