---
tag: spec-governance
author: jay-flowers
category: gotcha
created_at: 2026-11-05T15:18:11Z
identity: spec-governance-20261105T151811-jay-flowers
tier: draft
---

Treat a requirement-id citation in a comment on the code you are editing as a live contract pointer, not history. PR 957 edited scaffold.go's opencodeGitignore constant — whose comment cites "SC-FR-015, design D9" — to add package-lock.json as a fifth canonical .opencode/.gitignore line, while the unarchived init-override-stubs-gitignore-sync change's SC-FR-015 required exactly four canonical lines (node_modules, bun.lock, .vitest/, coverage/). The contradiction was invisible to the stack-reconciliation grep for the rewritten command ("npm ci") because the pinned artifact is a content list, not a command. When editing an asset whose comment cites a requirement id, open the owning unarchived change's delta and check its exactly-clauses and want-strings; stack-reconciliation greps must cover pinned canonical content, not just mechanism names.
