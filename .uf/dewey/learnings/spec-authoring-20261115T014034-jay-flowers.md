---
tag: spec-authoring
author: jay-flowers
category: pattern
created_at: 2026-11-15T01:40:34Z
identity: spec-authoring-20261115T014034-jay-flowers
tier: draft
as_of: "unbound-force@7f44cba6be6f30f6c3cec406daab342434cc1474"
---

A per-finding decision gate must state the post-decision control flow explicitly: which decision outcomes advance the gated step and which schedule the next review iteration. "Advance allowed when every finding has a decision" with apply in the decision set silently lets the gate close on consented fixes with no re-review — a review-policy change that contradicts the iteration semantics the change itself preserves. Specify the mapping (apply → iteration N+1 verification; override/pause → advance or stop) or require an approving verdict before advancement. Seen in harden-spec-review-gate HARDEN-002's "All findings decided — advance allowed" scenario vs HARDEN-003's next-iteration scenario.
