---
tag: instruction-quality
author: jay-flowers
category: gotcha
created_at: 2026-11-16T14:09:38Z
identity: instruction-quality-20261116T140938-jay-flowers
tier: draft
as_of: "unbound-force@c9e35b82b9b5a4718abfc2f4bd320eb849aa60c1"
---

Splice-editing new instruction into an existing em-dash chain degrades pronoun referents. When a DE-001/DE-005 fix inserts a tool-behavior aside (e.g., a tool-derived field like remaining_ms) between a sentence's subject ("The iteration count rides the free-form data parameter as the gate-distinct fixLoopPhase[N] key") and its closing clause ("-- so record it on every checkpoint"), every noun the aside introduces becomes a competing antecedent for the closing pronoun, and readers can act on the wrong one (recording remaining_ms instead of the iteration count, or reading a tool-derived field as a call input). Restructure into separate sentences that name their referent — one instruction per sentence in command prose — rather than splicing into the existing chain.
