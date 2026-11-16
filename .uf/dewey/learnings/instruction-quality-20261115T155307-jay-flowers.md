---
tag: instruction-quality
author: jay-flowers
category: pattern
created_at: 2026-11-15T15:53:07Z
identity: instruction-quality-20261115T155307-jay-flowers
tier: draft
---

When instruction prose enumerates a tool's refusal outcomes ("any X refuses with A; any Y refuses with B"), state the precedence for inputs matching both conditions. In harden-spec-review-gate, evaluateSpecReviewGate checks pause before apply, so a mixed apply+pause decision set refuses with spec_review_gate_paused — prose stating the apply rule as universal ("any apply decision refuses with pending_approving_iteration") mis-states the control flow for exactly the mixed sets the Pause & exit option creates.
