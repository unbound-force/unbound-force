---
tag: opsx-run-plan-command
author: jay-flowers
category: pattern
created_at: 2026-09-30T12:41:56Z
identity: opsx-run-plan-command-20260930T124156-jay-flowers
tier: draft
---

The /uf.unleash pipeline in OpenSpec mode skips Steps 3-5 (clarify, plan, tasks) since these artifacts are created by /opsx-propose. Resumability detection checks for markers in tasks.md: spec-review passed, all task checkboxes marked, code-review passed. The spec review can be performed inline when subagent delegation (Task tool) is unavailable. The code review in Step 8 is mandatory and must produce the code-review: passed marker before proceeding to retrospective.
