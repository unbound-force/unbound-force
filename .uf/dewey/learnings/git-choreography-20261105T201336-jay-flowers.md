---
tag: git-choreography
author: jay-flowers
category: gotcha
created_at: 2026-11-05T20:13:36Z
identity: git-choreography-20261105T201336-jay-flowers
tier: draft
---

In command procedures, a tree-cleanliness gate written as `git status --porcelain --untracked-files=no` passes working trees that hold untracked files colliding with paths the target branch tracks; the subsequent `gh pr checkout` (or an equivalent `git checkout -B <branch> FETCH_HEAD`) then aborts with git's 'untracked working tree files would be overwritten by checkout'. The failure must be handled on the checkout itself: capture the error, STOP, and name the remediation (remove, rename, or `git stash -u` the colliding files — plain `git stash` skips untracked). A downstream branch/SHA verification STOP will still fire, but it mis-attributes the cause as unpushed commits or divergence and discards the real error. The uf.rebase-pr Phase 3 fallback (fetch head-repo URL + `git checkout -B FETCH_HEAD`) does not resolve this failure mode — its documented cause is head-branch deletion (remote resolution), and the manual checkout trips the same git guard. Verified in PR #955 feedback Item 5 against .opencode/commands/uf.address-feedback.md Section 1.2.1 and uf.rebase-pr.md:96-103.
