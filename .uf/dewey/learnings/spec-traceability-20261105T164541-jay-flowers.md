---
tag: spec-traceability
author: jay-flowers
category: pattern
created_at: 2026-11-05T16:45:41Z
identity: spec-traceability-20261105T164541-jay-flowers
tier: draft
---

When a change reclassifies a scaffold path from tool-owned to user-owned, extend the repo's established user-owned survival test pattern to it: pre-seed a foreign file in a t.TempDir() target, run Run(), assert the file survives byte-identical. Predicate-flip assertions (isToolOwned/isDivisorAsset tables) plus command-string pins (e.g. --no-package-lock) verify the mechanism but leave the spec scenario's observable outcome ("the next uf init neither overwrites nor reads it") unverified at the composition level. Observed in PR 957 (stop-shipping-plugin-lock): the lock moved to the user-owned class alongside custom agents, matrix overrides, and custom packs — which all have composed survival tests — but the lock did not get one until reviewer feedback flagged the traceability gap.
