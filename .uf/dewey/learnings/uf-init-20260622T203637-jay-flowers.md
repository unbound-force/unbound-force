---
tag: uf-init
author: jay-flowers
category: pattern
created_at: 2026-06-22T20:36:37Z
identity: uf-init-20260622T203637-jay-flowers
tier: draft
---

Cross-repo comparison between unbound-force and downstream hero repos (like gaze) is an effective technique for discovering gaps in slash command customization. The /uf-init command applies customizations to OpenSpec skill files and speckit command files, but its semantic idempotency checks were too coarse -- detecting a basic branch check and concluding the entire branch enforcement category was present. Splitting idempotency checks into independent variant checks (basic branch check, dirty tree check, commit-before-archive, branch-switch confirmation) with specific marker strings prevents partial application from being mistaken for complete application. Each variant uses a different structural marker: opsx/<name> for basic check, git status --short for dirty tree, git add/git commit for commit-before-archive, and uncommitted changes for branch-switch confirmation.
