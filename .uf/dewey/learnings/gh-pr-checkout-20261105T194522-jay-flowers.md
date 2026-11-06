---
tag: gh-pr-checkout
author: jay-flowers
category: gotcha
created_at: 2026-11-05T19:45:22Z
identity: gh-pr-checkout-20261105T194522-jay-flowers
tier: draft
---

gh pr checkout succeeds for fork PRs even when "Allow edits by maintainers" is off, and in that case it configures branch.<b>.remote to the base repository and branch.<b>.merge to the read-only refs/pull/<N>/head ref. A fully-configured branch.<b>.remote/.merge pair is therefore an existence signal, not a writability signal. Any workflow that plans to push the PR head after gh pr checkout must validate the merge-ref shape (a real branch ref, refs/heads/...) and stop early on refs/pull/ — otherwise the only failure point is the final git push, after all upstream work (assessment, triage, commits, review re-runs) has already run, and standard "retry the push" guidance is futile because GitHub rejects pushes to refs/pull/* unconditionally. Found in the /uf.address-feedback Section 1.2.1 review (PR #955, feedback item 2).
