---
tag: documentation-pipeline
author: jay-flowers
category: pattern
created_at: 2026-11-07T16:39:07Z
identity: documentation-pipeline-20261107T163907-jay-flowers
tier: draft
---

Remediation strings are executable user-facing guidance: when the remediated command changes, re-validate the remediation against the new command's failure modes on every runtime the version gate accepts. In PR 957, npm ci became a lockless npm install, and the modified remediation constants in both uf init and uf doctor told users to re-run the exact command that crashes npm 10.9.9's Arborist when devDependencies are present — a runtime the doctor reports healthy. A remediation that re-runs the crashing command converts a recoverable failure into a support dead-end. Keep remediation constants in lockstep with the known-failure matrix of the command they name, and surface the escape hatch (the pinned npm 11 runner) in user-facing guidance when one exists.
