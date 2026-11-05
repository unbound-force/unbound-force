---
tag: error-handling
author: jay-flowers
category: gotcha
created_at: 2026-11-05T20:13:30Z
identity: error-handling-20261105T201330-jay-flowers
tier: draft
---

In command-file git choreography, a later verification step is a backstop, not error handling. uf.address-feedback Section 1.2.1 (PR #955) prescribed `gh pr checkout` after a `git status --porcelain --untracked-files=no` dirty gate: untracked files pass the gate, but they abort the checkout when they collide with files the PR branch tracks, and the only handler — the post-checkout branch/SHA mismatch STOP — misattributed the cause as local divergence. Two rules generalize: (1) every fallible command added to command prose needs its own failure contract (STOP with the specific error, or a documented fallback) — a downstream mismatch check stops the flow but reports the wrong cause; (2) before reusing a sibling command's fallback, check it against both the failure mode and downstream dependencies: uf.rebase-pr's `git fetch <head-repo-url> <head-ref>` + `git checkout -B <HEAD_BRANCH> FETCH_HEAD` fallback targets gh-layer failures (e.g. deleted head branch), does not resolve the untracked collision (git aborts identically), and configures no branch.<ref>.remote/.merge tracking — which push steps that read HEAD_REMOTE/HEAD_MERGE_REF from git config require.
