---
tag: command-authoring
author: jay-flowers
category: gotcha
created_at: 2026-11-05T20:13:34Z
identity: command-authoring-20261105T201334-jay-flowers
tier: draft
---

When command prose prescribes a state-changing git command (`gh pr checkout`, `git fetch`, `git push`), the same increment must prescribe its failure path — an explicit STOP that surfaces the command's exact error output. Gotcha: analyze the preceding guard and the guarded command as a pair. `git status --porcelain --untracked-files=no` deliberately ignores untracked files, so it passes exactly when an untracked file colliding with a tracked file on the target branch will abort `gh pr checkout` ("untracked working tree files would be overwritten by checkout"). A downstream branch/SHA re-check still stops the flow but misdiagnoses the cause because the checkout's real error was never captured. In-repo precedent: uf.rebase-pr.md Phase 3 (lines 96-103) documents a fetch-and-checkout fallback for a failed `gh pr checkout`; uf.address-feedback Section 1.2.1 initially shipped without checkout-failure handling (PR #955, feedback item 5).
