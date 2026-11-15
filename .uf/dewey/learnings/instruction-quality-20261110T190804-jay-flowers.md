---
tag: instruction-quality
author: jay-flowers
category: context
created_at: 2026-11-10T19:08:04Z
identity: instruction-quality-20261110T190804-jay-flowers
tier: draft
---

Code Review Mode scope gate for the Linguist: a diff touching only JSON package manifests (.opencode/package.json and its scaffold mirror internal/scaffold/assets/opencode/package.json), provenance testdata, CHANGELOG.md, docs/, or Go/TypeScript source names no instruction file — none match the instruction-files table (.opencode/agents|commands|skills|uf/packs *.md, AGENTS.md, internal/scaffold/assets/opencode/**/*.md; the scaffold-mirror pattern requires .md, so the JSON manifest mirror is out). The Linguist skips Step 1 (instruction-lint) and Step 2 (clarity) per the "diff names no instruction file" rule and returns APPROVE with zero findings. Judging Gatekeeping compliance of the dependency pins themselves (pinned-version updates, override provenance) is the Guard's domain; docs prose quality is the Scribe/Curator's.
