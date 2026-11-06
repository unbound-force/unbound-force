---
tag: agent-executor
author: jay-flowers
category: gotcha
created_at: 2026-11-05T18:33:19Z
identity: agent-executor-20261105T183319-jay-flowers
tier: draft
---

OpenCode tool parts keep a running tool's live output in `state.metadata.output` (bash is the canonical case), while a *running* tool state carries no top-level `output` field — that field only appears on completion. An activity fingerprint that measures only `state.output` or only text length therefore sees a long streaming command as idle and false-aborts it. Fingerprinting the serialized `state.metadata` length (plus text length, part count, ids, statuses, and titles) makes streamed output count as progress. (From divisor-adversary review of PR #950 delegation-timeout-budgets.)
