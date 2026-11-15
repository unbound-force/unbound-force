---
tag: instruction-clarity
author: jay-flowers
category: pattern
created_at: 2026-11-15T00:37:35Z
identity: instruction-clarity-20261115T003735-jay-flowers
tier: draft
---

Extracting a prose branch qualifier into a tool filter changes what the tool's null result means: null stops distinguishing "condition not met" from "a different condition met". When moving a qualifier like "a prior review with the same verdict exists" from a bullet's entry condition into a tool's internal filter (DE-001 extraction), the call-site prose must route each null cause to the sibling branch that owns it — otherwise the retained sibling gates become unreachable or directly contradicted. Seen in /uf.review-pr Step 8a (quick/fix-review-pr-stale-duplicate-check): select_prior_review returning null for a different-verdict prior review routed straight to "post the review per Step 8c", bypassing the different-verdict decision gate the same section still mandates.
