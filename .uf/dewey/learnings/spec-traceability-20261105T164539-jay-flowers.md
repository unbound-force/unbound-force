---
tag: spec-traceability
author: jay-flowers
category: pattern
created_at: 2026-11-05T16:45:39Z
identity: spec-traceability-20261105T164539-jay-flowers
tier: draft
---

OpenSpec delta scenarios that end with an observable-outcome clause (e.g., "the next uf init neither overwrites nor reads it") need one composed acceptance test exercising the outcome end-to-end. Piecewise assertions of the underlying mechanisms — ownership predicate flips, gitignore lines, install command flags — verify each cause independently but leave the scenario clause itself untraceable to any test. Verified on PR #957 (stop-shipping-plugin-lock): the SC-FR-003 foreign-lock scenario's final clause had zero composed coverage while all mechanisms were separately tested, and no base-era test could have covered it because the clause contradicted base behavior (the lock was then a tool-owned embedded asset). The composed test is cheap when Run() is injectable: write foreign bytes into the target before Run(), run with faked activation commands, assert byte-identical survival.
