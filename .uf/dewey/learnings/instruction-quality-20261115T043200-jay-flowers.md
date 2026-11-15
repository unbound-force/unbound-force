---
tag: instruction-quality
author: jay-flowers
category: gotcha
created_at: 2026-11-15T04:32:00Z
identity: instruction-quality-20261115T043200-jay-flowers
tier: draft
---

When editing instruction files (commands, agents, skills, packs), a newly added word can collide with an established same-file sense. In .opencode/commands/uf.review-council.md, adding "non-empty truncated manifest-mode text" (finish-reason `length` sense) collided with the file's Section 4 "Truncation recovery" (host-side tool-result truncation sense) — the same word now names two different mechanisms in one load set. Before introducing a content word in a new sense, search the file for its existing usage and qualify the new sense at introduction (e.g., "a run that ended at the `length` finish reason") so every reading resolves the same way.
