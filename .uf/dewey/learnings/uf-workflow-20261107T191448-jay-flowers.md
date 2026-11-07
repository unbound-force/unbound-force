---
tag: uf-workflow
author: jay-flowers
category: pattern
created_at: 2026-11-07T19:14:48Z
identity: uf-workflow-20261107T191448-jay-flowers
tier: draft
---

Post-push follow-up steps must not fail the push call: the primary operation already landed, so a failure envelope invites a redundant — possibly force — retry. PR 948's git_push demonstrates the correct shape: when `git branch --set-upstream-to=<remote>/<branch>` fails after a landed push, the tool still returns success with a structured `warning` field carrying the exact remediation command, and the structured result keeps old_sha/new_sha/updated truthful about what actually happened (Constitution III observable quality; design D3 of retroactive-push-skill-contracts). The symmetric rule for migrations: a removal that fails must warn and keep the file rather than count it as migrated — never report partial success as complete.
