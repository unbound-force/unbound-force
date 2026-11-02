---
tag: spec-governance
author: jay-flowers
category: gotcha
created_at: 2026-11-01T19:23:31Z
identity: spec-governance-20261101T192331-jay-flowers
tier: draft
---

When a fix-loop proposal acknowledges a workflow deviation and defers its normalization to "a separate future governance change, out of scope for this fix loop", the deferral must be captured as a tracked issue referenced from the proposal or PR body before merge — an in-proposal prose promise has no verifiable existence at review time and tends to evaporate after the branch merges. Reviewer verification step: when a proposal defers work as out-of-scope, check the PR body's linked issues and the change's cited issues for the promised follow-up; if nothing is referenced, flag the untracked deferral (bookkeeping gap, non-blocking). Observed on quick/accept-variantless-parent-model (PR #898): the quick-branch-carrying-OpenSpec-artifacts divergence is documented at proposal.md line 47 as future governance work, but the change cites only documentation issues (#891, website#360) and no governance-normalization issue is referenced anywhere in the diff — the prior council's advisory follow-up cannot be confirmed as tracked.
