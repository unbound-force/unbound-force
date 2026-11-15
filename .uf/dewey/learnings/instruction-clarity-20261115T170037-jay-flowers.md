---
tag: instruction-clarity
author: jay-flowers
category: gotcha
created_at: 2026-11-15T17:00:37Z
identity: instruction-clarity-20261115T170037-jay-flowers
tier: draft
---

When command prose enumerates a tool's refusal outcomes, mixed input sets must be checked against the tool's actual evaluation order. Listing "any apply decision refuses with X ... any pause decision refuses with Y" predicts the wrong refusal for a set containing both whenever the tool evaluates Y first (evaluateSpecReviewGate checks pause before apply). Routing keyed to the observed refusal limits the damage, but the prose mispredicts the mixed case and agents may pre-plan or even reshape inputs to match the predicted path. Either state the precedence ("a pause outranks an apply in a mixed set") or enumerate outcomes only for the single-kind cases the tool guarantees.
