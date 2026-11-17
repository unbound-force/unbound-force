---
tag: review-scope
author: jay-flowers
category: gotcha
created_at: 2026-11-15T00:34:11Z
identity: review-scope-20261115T003411-jay-flowers
tier: draft
---

Linguist Spec Review Mode scope gate: an OpenSpec change that touches only TypeScript tool code under `.opencode/` (lib/*.ts, plugins/*/index.ts), its `internal/scaffold/assets/opencode/` .ts mirrors, tests, eval fixtures, and CHANGELOG.md proposes no instruction-file changes. The `.opencode/` prefix alone does not make a path an instruction file — only the .md prose paths in the instruction-files table count (`.opencode/agents/*.md`, `.opencode/commands/*.md`, `.opencode/skills/*/SKILL.md`, `.opencode/uf/packs/*.md`, `AGENTS.md`, `internal/scaffold/assets/opencode/**/*.md`). The correct Linguist verdict for such a change is a no-findings APPROVE; the tool code belongs to the Determinist/Architect/Tester review slice, and spec-artifact structure (delta shape, numbering allocation) is the Determinist's "planned deltas" checklist.
