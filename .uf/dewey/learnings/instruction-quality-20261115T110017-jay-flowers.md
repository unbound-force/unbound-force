---
tag: instruction-quality
author: jay-flowers
category: pattern
created_at: 2026-11-15T11:00:17Z
identity: instruction-quality-20261115T110017-jay-flowers
tier: draft
---

Linguist Code Review Mode scope is path-based, not content-based: when a diff's changed paths name no instruction files (.opencode/agents|commands|skills|uf/packs .md, AGENTS.md, or .md mirrors under internal/scaffold/assets/opencode/), the correct outcome is zero findings and APPROVE with no diff retrieval. Scaffold mirrors count as instruction files only when they are .md files — .ts/.yaml/.json mirrors under internal/scaffold/assets/ are out of scope, as are docs/, README.md files (including schema READMEs), openspec/ and specs/ artifacts, and all TypeScript/Go source. A pure TypeScript plugin/lib + schema/config diff (e.g., the compaction-storm change) therefore yields an empty Linguist plate by design; the pre-flight instruction-lint "SKIP, no in-scope files" row corroborates the determination.
