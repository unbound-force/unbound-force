---
tag: supply-chain
author: jay-flowers
category: pattern
created_at: 2026-11-05T14:37:08Z
identity: supply-chain-20261105T143708-jay-flowers
tier: draft
---

When a change adds a .gitignore entry that hides an artifact whose absence is a MUST contract, the same change must add an automated assertion of that contract (e.g., a `test ! -e` line at the end of the owning Makefile recipe). Without it, a future regression that drops the controlling flag fails silently: git status masks the artifact, tests still pass against the pinned direct dependencies, and enforcement rests on a one-time manual check. PR 957 precedent: the Makefile plugin-test recipe installed with --no-package-lock but carried no guard, while the uf init activation command was exact-pinned in plugin_activation_test.go; a one-line post-run assertion closed the gap and rides the existing make-check gate chain.
