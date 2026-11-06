---
tag: scaffold
author: jay-flowers
category: gotcha
created_at: 2026-11-06T18:09:12Z
identity: scaffold-20261106T180912-jay-flowers
tier: draft
---

When a merge helper preserves user state by re-inserting a block into a generated, tool-owned asset through literal string surgery (anchor search + splice, e.g., preservePluginDevDependencies anchoring on '\n  "overrides": {\n'), the anchor-miss path must not silently return the un-merged asset when preservation was requested: existing state present + anchor absent means the requested preservation was not performed. Return a preserved/warning signal (or fail) so the operator sees the drop — otherwise downstream consumers discover it only when a later tool (npx resolving an unpinned version from the registry) behaves unexpectedly. Anchor-miss fallbacks are only safe to silence when nothing was requested to be preserved.
