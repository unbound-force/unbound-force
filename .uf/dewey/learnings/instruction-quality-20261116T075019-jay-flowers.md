---
tag: instruction-quality
author: jay-flowers
category: gotcha
created_at: 2026-11-16T07:50:19Z
identity: instruction-quality-20261116T075019-jay-flowers
tier: draft
---

When a change introduces a new human-interaction category into pipeline instruction files, sweep every exhaustive enumeration of interactions in the load set, not only the normative ones. In the synthetic-compaction guard change (opsx/synthetic-compaction-continue-prompt-rec), iteration 1 caught the two normative enumerations ("the only legitimate mid-pipeline human interactions" in uf.unleash.md, "question tool use is limited to" in uf.run-pipeline-plan.md) and updated them; iteration 2 found a third, descriptive inventory in the same file ("The remaining human control points are...") still omitting the new category. Before declaring a consistency fix complete, grep the changed files and their siblings for "only legitimate", "limited to", "control points", "remaining", and "human interaction", and update or explicitly subsume every enumeration that names the concept.
