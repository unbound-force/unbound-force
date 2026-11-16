---
tag: instruction-quality
author: jay-flowers
category: gotcha
created_at: 2026-11-16T07:12:24Z
identity: instruction-quality-20261116T071224-jay-flowers
tier: draft
---

When an instruction change adds a new mandated human-interaction gate to a command or agent file, audit the same file (and every sibling carrying the same block) for exhaustive enumerations of legitimate interactions — "the only legitimate mid-pipeline human interactions are X and Y" or "question tool use is limited to X and Y". A new MUST appended elsewhere in the same section leaves a logical seam: post-compaction agents re-anchoring on partial context can resolve the stale "only" claim against the new gate. When a byte-exact pin test freezes the new prose, the fix belongs in the enumeration sentence, not the new block — rewording the pinned block breaks the pin.
