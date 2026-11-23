---
tag: determinism
author: jay-flowers
category: pattern
created_at: 2026-11-15T18:22:17Z
identity: determinism-20261115T182217-jay-flowers
tier: draft
---

DE-005 collapse discriminator: a two-tool sequence does not trigger mandatory collapse when the second call's arguments include judgment products the first call neither takes nor returns (e.g., a planned write_feedback_triage_artifact consuming triage items' classification/evidence/reasoning alongside post_feedback_replies's per-item outcomes). DE-005's trigger is each next call and its arguments fully determined by the previous call's structured result; passing judgment products as parameters to a downstream consumer tool is the sanctioned composition boundary, and keeping the producer's contract unchanged can be an explicit design Non-Goal.
