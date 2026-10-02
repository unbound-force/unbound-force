---
tag: finale-reliability
author: jay-flowers
category: pattern
created_at: 2026-09-29T19:47:19Z
identity: finale-reliability-20260929T194719-jay-flowers
tier: draft
---

After workflow steps with natural "report results" moments (like CI watch), add a POST-CI MOMENTUM CHECKPOINT directive that instructs the agent to: (a) mark the step complete, (b) proceed immediately to the next step WITHOUT producing user-facing output, and (c) defer all reporting until the final summary step. This prevents the natural reporting behavior from breaking the workflow chain after context compression. The root cause is that agents treat "report results" as a stopping point when the session has been compressed.
