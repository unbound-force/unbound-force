---
tag: workflow-resume-gates
author: jay-flowers
category: gotcha
created_at: 2026-11-05T21:49:09Z
identity: workflow-resume-gates-20261105T214909-jay-flowers
tier: draft
---

An ancestry-based re-entry gate (`git merge-base --is-ancestor <frozen_sha> HEAD`) verifies commit containment, not branch identity: any branch that contains the frozen SHA passes — including main after the PR merged upstream, or a local branch that merged the PR head for testing. When the accepting command later records its push target from `git config branch.<name>.remote/.merge` and gates only on config presence plus a `refs/pull/` prefix, the wrong-branch state flows through every phase to `git push <remote> HEAD:<merge_ref>` — a direct push to the wrong ref (e.g., main), violating branch protection and stranding commits on the wrong branch. The descendant-acceptance pattern (correct for unpushed-commit resume, per the checkpoint-recovery learning) must be paired with push-target validation against the frozen PR context: HEAD_MERGE_REF must equal `refs/heads/<head_ref>` AND HEAD_REMOTE must resolve to the PR's head repository — the remote check is load-bearing because a fork PR whose head is a default branch named `main` produces a merge ref indistinguishable from the operator's own main by ref name alone. Found reviewing /uf.address-feedback Section 1.2.1 (PR #955): the open-PR check in Section 1.1 is skipped when PR_NUMBER is passed explicitly, so a merged-PR re-run from main reaches the descendant bullet directly.
