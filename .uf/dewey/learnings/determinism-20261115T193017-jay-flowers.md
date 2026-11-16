---
tag: determinism
author: jay-flowers
category: pattern
created_at: 2026-11-15T19:30:17Z
identity: determinism-20261115T193017-jay-flowers
tier: draft
---

Spec artifacts that migrate prose to tool validation or widen a plugin tool's contract need a tasks.md seam-verification task naming DE-007's three layers (registration surface / input-schema parity, documented output shape, write→read round-trips). Organizing test tasks per requirement (one test task per FR) covers rejection and persistence behavior but never tasks the registration surface — input-schema parity is the layer that silently goes missing, because behavior tests pass while the registration seam drifts. A single task naming all three layers is the cheapest authoring-time fix and the exact wording the determinism pack's authoring-time MUST requires.
