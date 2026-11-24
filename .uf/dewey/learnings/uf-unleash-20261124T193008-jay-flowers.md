---
tag: uf-unleash
author: jay-flowers
category: gotcha
created_at: 2026-11-24T19:30:08Z
identity: uf-unleash-20261124T193008-jay-flowers
tier: draft
as_of: "unbound-force@12bb5f8799eea38393959349f27354e1bb031a77"
evidence:
  - ".opencode/agents/gaze-reporter.md"
---

In parented child sessions running at subagent depth 1, invoking the gaze-reporter agent through the Task tool fails with 'Subagent depth limit reached (1)'. The working fallback used for the /uf.unleash Step 6 Phase 1b Gaze report on 2026-11-24: read the agent definition (.opencode/agents/gaze-reporter.md), which documents its exact underlying command, and run that command directly via bash (gaze report --format=markdown ./... for full mode), capturing stdout as the informational report and recording the mechanism deviation in the dispatch review context. Phase 1b remains conditional and non-blocking, so a failed gaze command itself still allows the council dispatch to proceed.
