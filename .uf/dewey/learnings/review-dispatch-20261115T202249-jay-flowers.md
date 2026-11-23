---
tag: review-dispatch
author: jay-flowers
category: gotcha
created_at: 2026-11-15T20:22:49Z
identity: review-dispatch-20261115T202249-jay-flowers
tier: draft
---

The spec-review iteration cap in the review-dispatch plugin (SPEC_REVIEW_ITERATION_CAP = 3, i.e. initial review plus 3 fix iterations, enforced by specReviewDispatchGuard in .opencode/plugins/review-dispatch/index.ts) refuses a 5th specs-mode dispatch with spec_review_iteration_cap_exceeded (retryable: false) even when the latest iteration is fully adjudicated with apply decisions and the consented fixes have already been applied to the spec artifacts. The guard's missing_decisions and paused refusals clear once decisions are recorded, but the cap check runs unconditionally afterward, so a pending_approving_iteration gate state cannot be certified once 4 REQUEST CHANGES iterations are recorded on the branch. There is no override parameter, config knob, or env var — the record lives in .uf/progress state under the sanitized branch key (command uf.unleash, data key spec-review) and is reconciled from dispatch artifacts, so clearing or editing state to reset the count is gate tampering (CRITICAL). The only sanctioned exits are the tool-rendered spec-review paused path (format_unleash_exit variant spec-review) or a deliberate code change to the cap through normal review. Encountered on opsx/feedback-triage-artifact-tool (issue #1022) after the human adjudicated all 8 iteration-4 findings as apply and directed a re-run.
