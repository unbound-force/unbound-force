---
tag: instruction-file-hygiene
author: jay-flowers
category: gotcha
created_at: 2026-11-03T19:14:53Z
identity: instruction-file-hygiene-20261103T191453-jay-flowers
tier: draft
---

When editing hand-wrapped instruction files (SKILL.md phases, command markdown, tool-description string concatenations), re-wrap the whole edited paragraph to the file's existing column discipline instead of appending text to an already-wrapped line. In the pre-flight SKILL.md (uniform <=72-character prose wrap across all 589 pre-change lines), appending one clause to a mid-paragraph line produced a ~94-character line; the same edit pattern pushed two uf-workflow tool-description strings past their block's 
