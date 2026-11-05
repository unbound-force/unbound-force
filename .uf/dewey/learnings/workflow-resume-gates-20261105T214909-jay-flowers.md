---
tag: workflow-resume-gates
author: jay-flowers
category: gotcha
created_at: 2026-11-05T21:49:09Z
identity: workflow-resume-gates-20261105T214909-jay-flowers
tier: draft
---

An ancestry-based re-entry gate (`git merge-base --is-ancestor <frozen_sha> HEAD`) verifies commit containment, not branch identity: any branch that contains the frozen SHA passes — including main after the PR merged upstream, or a local branch that merged the PR head for testing. When the accepting command later records its push target from `git config branch.<name>.remote/.merge` and gates only on config presence plus a `refs/pull/` prefix, the wrong-branch state flows through every phase to `git push <re
