---
tag: error-handling
author: jay-flowers
category: gotcha
created_at: 2026-11-05T20:13:32Z
identity: error-handling-20261105T201332-jay-flowers
tier: draft
---

In unbound-force command prose, every fallible external command carries a point-of-failure prescribed path (surface the specific error, STOP, name the remedy) — see uf.address-feedback.md line 222 ("On API failure ... report the specific error and STOP") and line 1076 ("If push fails ... preserve local commits"). A dirty check scoped `git status --porcelain --untracked-files=no` does not protect a later `gh pr checkout`: git aborts the checkout when an untracked working-tree file would be overwritten by the target branch, so the checkout step needs its own fail-closed branch. uf.rebase-pr.md Phase 3 (lines 96-103) documents a fetch-and-checkout fallback for a failed `gh pr checkout`, but reusing it in /uf.address-feedback would leave branch.<head_ref>.remote/.merge unset (a URL-based fetch plus `git checkout -B <HEAD_BRANCH> FETCH_HEAD` configures no push target) and trip the later HEAD_REMOTE/HEAD_MERGE_REF STOP — the minimal consistent fix is an explicit STOP at the checkout, not the fallback.
