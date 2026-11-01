---
tag: spec-governance
author: jay-flowers
category: pattern
created_at: 2026-11-01T17:49:40Z
identity: spec-governance-20261101T174940-jay-flowers
tier: draft
---

When an OpenSpec delta claims same-name supersession with an archive-ordering constraint, verify the constraint is COMPLETE by grepping the requirement name across every unarchived change — not just the cited predecessor. An unnamed sibling delta carrying the same requirement name (especially another MODIFIED entry) would replace the corrected text at its own later archive, silently regressing the canonical spec; the ordering constraint must name every same-name delta or the supersession claim is incomplete. Verified on accept-variantless-parent-model (quick/accept-variantless-parent-model branch, 2026-11-01): grepping 'Requirement: Review Dispatch Artifact' across openspec/changes/ returned exactly two hits — the fanout ADDED (line 486) and this change's MODIFIED (line 116) — confirming the single-predecessor constraint; grepping 'Finalization Model Validation' returned the fallback's orphan MODIFIED only, with no ADDED anywhere, confirming the ADDED-reshape premise. This extends same-name matching: a name match against the cited target is necessary but not sufficient — the full predecessor set must also be enumerated.
