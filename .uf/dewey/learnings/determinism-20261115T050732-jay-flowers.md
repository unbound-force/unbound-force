---
tag: determinism
author: jay-flowers
category: gotcha
created_at: 2026-11-15T05:07:32Z
identity: determinism-20261115T050732-jay-flowers
tier: draft
---

Hand-synced prose allowlists of tool-emitted enumerations diverge silently. PR #1087 added the child_empty_response warning kind to consolidate_dispatch and had to manually add the same kind to the spill-validation allowlist in .opencode/commands/uf.review-council.md (plus its scaffold mirror) — nothing fails when the two lists drift apart; a missed entry silently discards valid warnings during truncation recovery. When a tool gains an enumerated output value that instruction prose validates, either migrate the validation into the tool (a spill-recovery tool owning the kind enumeration) or add a seam test/linter check that pins the prose list to the tool's enumeration.
