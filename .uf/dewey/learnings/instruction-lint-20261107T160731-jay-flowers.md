---
tag: instruction-lint
author: jay-flowers
category: pattern
created_at: 2026-11-07T16:07:31Z
identity: instruction-lint-20261107T160731-jay-flowers
tier: draft
---

A lint CLI that exits 2 when a --only path matches no lintable file (typo protection, so a misspelled path can never silently filter findings away) composes badly with reviewer personas fed from a diff's changed-file list: deleted and renamed-away paths are legitimately "changed" but match no existing artifact, so a single deletion converts the whole scoped run into the tool-failure lane — the persona reports a spurious could-not-run finding and skips its deterministic evidence for every other changed file in the change. When wiring a diff-driven file list into an existence-strict CLI filter, scope the list to paths that still exist in the working tree (deleted files cannot carry findings) and surface deletions through a separate judgment step (e.g. reference-rot review) instead of the tool-failure path.
