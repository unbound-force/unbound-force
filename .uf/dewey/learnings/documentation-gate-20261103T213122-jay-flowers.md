---
tag: documentation-gate
author: jay-flowers
category: gotcha
created_at: 2026-11-03T21:31:22Z
identity: documentation-gate-20261103T213122-jay-flowers
tier: draft
---

In unbound-force/unbound-force, user-facing changes must satisfy two documentation-issue gates that name different targets: repo AGENTS.md's Documentation gate requires an issue against the current repo before PR merge, while the org constitution's Cross-Repo Documentation rule — operationalized in the website repo's own AGENTS.md Website Documentation Gate — requires an issue in unbound-force/website before merge (explicitly covering breaking changes to any user-facing workflow). Established practice files
