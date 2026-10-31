---
tag: spec-governance
author: jay-flowers
category: gotcha
created_at: 2026-10-31T22:10:06Z
identity: spec-governance-20261031T221006-jay-flowers
tier: draft
---

OpenSpec supersession-by-new-delta is declarative, not mechanical. When multiple unarchived changes carry MODIFIED deltas for the same requirement with contradictory normative text, archive-time spec sync applies each delta at that change's own archive (agent-driven, per the openspec-archive-change workflow) with no cross-change ordering guard — so the corrected text survives only if its change archives last. A delta that claims to supersede older unarchived deltas should state the archive-ordering requirement explicitly (e.g., "this change MUST archive after X and Y") or the stale deltas should be reconciled; otherwise a later archive of an older change silently regresses the canonical spec to the superseded text. Verified in unbound-force review-dispatch: both advisor-host-model-fallback and review-council-multi-model-fanout carry over-strict both-non-null 'Finalization Model Validation' text while accept-variantless-parent-model corrects it, and no synced openspec/specs/review-dispatch/ exists to anchor the merge.
