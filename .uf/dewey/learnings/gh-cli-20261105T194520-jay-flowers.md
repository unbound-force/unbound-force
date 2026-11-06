---
tag: gh-cli
author: jay-flowers
category: gotcha
created_at: 2026-11-05T19:45:20Z
identity: gh-cli-20261105T194520-jay-flowers
tier: draft
---

`gh pr checkout` on a fork PR with "Allow edits by maintainers" disabled configures `branch.<b>.remote` to the base repository and `branch.<b>.merge` to `refs/pull/<N>/head` — GitHub's read-only pull ref. Both git-config values are set, so presence-based validation passes, but `git push` to any `refs/pull/*` ref is always rejected by GitHub ("deny updating a hidden ref"), and `git fetch` of the same ref succeeds, so nothing fails until push time. Any UF command that records a push target from `gh pr checkout` must reject merge refs carrying the `refs/pull/` prefix at the gate — before assess, commits, and review dispatches run — instead of failing after the full pipeline has executed. Related precedent: `uf.rebase-pr.md` Phase 3 carries a `gh pr checkout` failure fallback, and the port-uf-rebase-pr design (D3) verifies the branch's upstream resolves to the intended head remote/branch before pushing for exactly this reason.
