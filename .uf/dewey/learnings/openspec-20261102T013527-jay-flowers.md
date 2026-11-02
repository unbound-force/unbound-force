---
tag: openspec
author: jay-flowers
category: gotcha
created_at: 2026-11-02T01:35:27Z
identity: openspec-20261102T013527-jay-flowers
tier: draft
---

When an OpenSpec change reverses or amends a behavioral contract, build the amendment inventory from the affected function's full requirement lineage, not just the numbered sibling lineages in the same capability family: grep the function name (e.g. ensureGitignore) across ALL active and archived change specs before writing the "Amendment of prior decisions" section. Verified during the init-override-stubs-gitignore-sync spec review (2026-11-01): the change's SC-FR-014 reversed the archived gitignore capability's "marker present → skip the append" MUST (2026-04-28-gitignore-init, specs/gitignore.md:15-17, scenario at 38-43) while the amendment section meticulously covered SC-FR-010 and D10 — the un-numbered capability owned by a separate archived change was invisible to the numbered-lineage (MP-014) scan. Un-numbered requirements in older capability specs are the blind spot: numbering headers record SC-FR-xxx allocations, but nothing in the delta discipline forces a function's older contracts into amendment scope, so the lineage accumulates contradictory MUSTs (gitignore capability: "not modified"; scaffold SC-FR-014: "rewritten in place"). Detection recipe for spec review: for each MUST whose behavior changes, grep the implementing function name across openspec/changes/**/specs/ (including archive/) and check every hit's requirement for a MODIFIED delta or amendment entry in the change under review.
