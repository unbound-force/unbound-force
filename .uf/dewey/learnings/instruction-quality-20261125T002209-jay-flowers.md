---
tag: instruction-quality
author: jay-flowers
category: pattern
created_at: 2026-11-25T00:22:09Z
identity: instruction-quality-20261125T002209-jay-flowers
tier: draft
as_of: "unbound-force@c9241969995f590083f3f56f0a0347168c4d56f9"
---

Rewriting a shared skill's protocol from caller-executed steps to tool-owned acquisition must, in the same change, update: (a) every consumer command the skill's own Consumers table names that still prescribes the old procedure, (b) consumers that rely on the capability on paths where the new owner never runs (feedback-mode dispatches, non-dispatch tiers), and (c) the skill's own Result Format sections the old procedure populated. Miss any one and unchanged consumers instruct agents to execute a deleted procedure or fabricate output sections with no producer — IQ-10 contradictions anchored in the changed protocol. Seen in intent-satisfaction-check: review-context Protocol 2 moved to dispatch-owned acquisition and left uf.review-pr 3.8b (prescribes the deleted parse/validate/fetch/extract procedure), uf.address-feedback 2.1.5 (Tier 1 non-dispatch path with no acquisition owner), and the skill's Linked Issues Result Format section dangling.
