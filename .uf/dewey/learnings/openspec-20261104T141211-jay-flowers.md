---
tag: openspec
author: jay-flowers
category: gotcha
created_at: 2026-11-04T14:12:11Z
identity: openspec-20261104T141211-jay-flowers
tier: draft
---

OpenSpec changes that grow a review-follow-ups phase superseding an earlier task's rule must reconcile every artifact that states the rule, or first-revision residue survives at multiple sites. Verified on PR 949 (review-dispatch-consumer-bounds): the proposal-dedup rule switched from overlap-coefficient >= 0.5 with a 4-token floor to Jaccard >= 0.3 with a 6-token floor, and five residue sites survived the switch — task 1.2 still describing the old algorithm with no supersession cross-reference, phase headings numbered 1,2,3,5,4 (MP-018), design decisions ordered D1-D5,D7,D6, a test name saying "fewer than four tokens", and RD-FR-055's "Distinct lessons are kept" GIVEN still framed as "fewer than half of the smaller proposal's tokens", which no longer implies non-merge under Jaccard >= 0.3 (100-token sets sharing 49 tokens score ~0.32 and merge). Reconciliation checklist: add a supersession cross-reference on the superseded task (design.md's "The first revision used X... Review showed..." pattern), renumber phases into executed order, reposition or renumber appended decisions, rename tests off the dead threshold, and reword requirement scenarios off the superseded measure.
