---
tag: determinism
author: jay-flowers
category: gotcha
created_at: 2026-11-15T21:09:19Z
identity: determinism-20261115T210919-jay-flowers
tier: draft
---

Contract tests that enforce a rule over instruction prose by scanning a hand-enumerated list of directories silently lose coverage when the instruction surface grows: the heredoc-ban test (TestCommandContracts_RejectShellHeredocPayloads) scanned only .opencode/commands/ and its scaffold mirror until issue #1047 / PR #1124 extended it to .opencode/agents/ and .opencode/skills/ (plus each mirror) — directories that carried the same executable bash instruction blocks all along (the gap was latent, verified zero live instances 2026-11-10). When a gate test bans a payload class across instruction directories, derive the scan roots from the instruction-surface definition or add a drift check pinning the expected root set, so a new or forgotten instruction directory cannot fall out of the gate.
