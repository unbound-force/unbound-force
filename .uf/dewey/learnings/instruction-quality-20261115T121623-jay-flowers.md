---
tag: instruction-quality
author: jay-flowers
category: pattern
created_at: 2026-11-15T12:16:23Z
identity: instruction-quality-20261115T121623-jay-flowers
tier: draft
---

An instruction branch that imports a sibling branch's protocol by reference ("treat it as X") does not automatically transfer X's exception structure: when the importing branch carves its own exception ("rather than routing to …"), readers cannot tell which imported sub-rules survive, and the repeated-failure escalation is the first casualty — one reader escalates per the imported rule, another loops on the exception. Restate the second-occurrence escalation inside the importing branch. Observed in the timeout.kind "storm" row of .opencode/commands/uf.run-pipeline-plan.md, which imports the "idle" row while excepting its blocked-item escalation, leaving a second storm failure undefined.
