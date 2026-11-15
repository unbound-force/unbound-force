---
tag: determinism
author: jay-flowers
category: pattern
created_at: 2026-11-10T22:16:05Z
identity: determinism-20261110T221605-jay-flowers
tier: draft
---

divisor-determinism's review surface is path-gated: instruction files (.opencode/agents|commands|skills|uf/packs, AGENTS.md) and tool code (.opencode/plugins, .opencode/lib). A docs-only PR (e.g., CHANGELOG.md + docs/*.md) therefore yields zero DE-001–DE-008 findings and an APPROVE, even when the triggering feedback item references an out-of-diff file (e.g., a test-data provenance record) whose substance belongs to Guard/Adversary domains — the confinement rule blocks out-of-diff findings and the persona's out-of-scope list blocks cross-domain ones. Confirm the gate deterministically rather than by reading: `instruction-lint --base <sha> --changed` returns delta.files: [] and `tool-inventory --seams --base <sha>` returns an empty seams list, which also discharges DE-004 and DE-007 without further reading.
