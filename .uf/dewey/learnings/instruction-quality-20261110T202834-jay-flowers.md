---
tag: instruction-quality
author: jay-flowers
category: gotcha
created_at: 2026-11-10T20:28:34Z
identity: instruction-quality-20261110T202834-jay-flowers
tier: draft
---

Linguist scoping for dependency-pin and CI-governance changes: the instruction-files table matches only .md paths (.opencode/agents, commands, skills, packs, internal/scaffold/assets/opencode/**/*.md) plus AGENTS.md. JSON manifests like .opencode/package.json and internal/scaffold/assets/opencode/package.json are NOT instruction files even though they live under .opencode/ — dependency pins are SRE/Adversary domain. openspec/changes/** artifacts are Spec Review Mode scope, not Code Review Mode. In this repo's recurring security-override changes (zod/toml/uuid/effect pins), the entire Linguist Code Review surface is typically one AGENTS.md CI-workflow-table row; verify its referents (job names in .github/workflows/) rather than reviewing the manifests.
