---
tag: command-authoring
author: jay-flowers
category: gotcha
created_at: 2026-11-05T22:51:13Z
identity: command-authoring-20261105T225113-jay-flowers
tier: draft
---

When a change commits a learning that documents a control as load-bearing, the same increment must be verified to implement that control in the prose it governs — restored learning text and command prose can diverge silently. Evidence: PR #955 iteration 3 — the committed learning workflow-resume-gates-20261105T214909 states "HEAD_REMOTE must resolve to the PR's head repository — the remote check is load-bearing because a fork PR whose head is a default branch named main produces a merge ref indistinguishable from the operator's own main by ref name alone," while uf.address-feedback Section 1.2.1's push-target cross-check validates only HEAD_MERGE_REF == refs/heads/<head_ref>. A re-run on a merge-commit-merged fork-default-branch PR (explicit PR number skips Section 1.1's state checks; gh pr view still serves merged PRs) from the operator's main passes every gate — descendant bullet, charset gate, refs/pull/ check, merge-ref name equality — and pushes HEAD:refs/heads/main to the operator's own remote instead of the PR head repository. Verification step: diff each committed learning's MUST-list against the shipped command prose before approving the increment; the learning store is a record of required controls, not evidence they shipped.
