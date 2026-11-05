---
tag: gh-pr-checkout
author: jay-flowers
category: gotcha
created_at: 2026-11-05T22:51:16Z
identity: gh-pr-checkout-20261105T225116-jay-flowers
tier: draft
---

Merge-ref equality binds a recorded push target to the PR head branch NAME, not to the PR head REPOSITORY. A local branch named exactly head_ref whose branch.<name>.remote points at a different repository carrying the same commits (an author's mirror remote, or an operator's same-named fork branch) passes every gate in /uf.address-feedback Section 1.2.1 — HEAD descends the frozen head_sha so no checkout runs, the charset gate passes, the merge ref is not refs/pull/, and HEAD_MERGE_REF == refs/heads/<head_ref> matches by name — so Section 4.4 pushes the PR's commits to the wrong repository, the push succeeds there, and the PR silently never updates. The Section 1.2 frozen context (pr_number, base/head ref+sha) carries no head-repository identity, so the cross-check has nothing to compare HEAD_REMOTE against; learning workflow-resume-gates-20261105T214909 already names the remote check as load-bearing. Fix shape: freeze the head repository in the immutable context and verify HEAD_REMOTE resolves to it (git remote get-url for named remotes, URL equality otherwise), or resolve the push target server-side in the #965 checkout_pr_head tool where gh already resolves the head repository.
