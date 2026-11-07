---
tag: openspec
author: jay-flowers
category: pattern
created_at: 2026-11-05T16:27:34Z
identity: openspec-20261105T162734-jay-flowers
tier: draft
---

Quick-tier changes that alter shipped plugin-tool contracts leave live unarchived OpenSpec deltas contradicting the implementation, because the quick tier ships no spec artifacts and nothing reconciles them automatically. Before merging such a change, reconcile the affected live deltas in the same PR: annotate the contradicted requirement text with superseded-by references pointing at a retroactive OpenSpec change that carries the new contract (the fix-tui-git-output-leak annotation precedent inside finale-token-reduction, and unleash-deterministic-tools task 9.3's discipline — annotate stale delta text with superseded-by references "so the main spec never lands text contradicting the merged contract"). Delta-level annotations matter because the contradicted requirements live in unarchived changes, not yet in canonical openspec/specs/; archive ordering otherwise determines whether stale MUST text lands in the canonical spec.
