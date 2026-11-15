---
tag: instruction-gates
author: jay-flowers
category: gotcha
created_at: 2026-11-15T20:47:03Z
identity: instruction-gates-20261115T204703-jay-flowers
tier: draft
---

Instruction prose carrying executable bash blocks lives in three canonical trees — .opencode/commands/, .opencode/agents/, and .opencode/skills/ (nested <name>/SKILL.md) — each with a byte-identical mirror under internal/scaffold/assets/opencode/. A prose-behavior gate that scans only one tree leaves latent gaps: the #1011 heredoc ban in internal/scaffold/command_contract_test.go originally covered commands only, and issue #1047 extended it after a review found agents and skills carried 41 unscanned bash blocks (divisor-curator composes and posts issue content). When adding or extending an instruction-file gate, enumerate every instruction directory and its scaffold mirror, and walk the tree recursively (markdownFilesUnder) because layouts differ: commands and agents sit flat, skills nest one level down.
