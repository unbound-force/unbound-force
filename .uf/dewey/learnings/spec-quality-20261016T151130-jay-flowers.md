---
tag: spec-quality
author: jay-flowers
category: gotcha
created_at: 2026-10-16T15:11:30Z
identity: spec-quality-20261016T151130-jay-flowers
tier: draft
---

When a change ships spec artifacts alongside co-authored Dewey learnings, reconcile contradictions between them before merge: design.md decision D3 in openspec/changes/fix-tool-args-schema-mismatch claimed "any drift would surface as a tsc/vitest failure," while the learning plugin-tool-schemas-20261016T132705 (from the spec review) documents that optional-to-required parameter drift is tsc-silent. Amend the design text or add a supersession note so future readers do not trust an overstated guard.
