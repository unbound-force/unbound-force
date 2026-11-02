---
tag: review-dispatch
author: jay-flowers
category: pattern
created_at: 2026-11-02T17:47:50Z
identity: review-dispatch-20261102T174750-jay-flowers
tier: draft
---

When a change introduces a new error code into a tool-result contract that feeds verdict projection or classification consumers, the security review must enumerate every code-matching consumer and execute its decomposition logic against the new code rather than trusting the design doc's claim. In split-host-resolution-failure-classes (PR #920), the new host_model_structural code interacts with availabilityError() in review-dispatch/index.ts, which splits the code on underscores and requires a provider/model/runtime token AND an availability token (unavailable/availability/timeout/unsupported) or a _not_found suffix. Verified by decomposition: ["host","model","structural"] has "model" but no availability token, so availabilityError returns false — a structural cause reads as a configuration problem, not an availability blocker, keeping noSuccessResult's UNAVAILABLE-vs-INCONCLUSIVE projection honest. The symmetric check also confirmed host_model_unavailable still classifies as availability (correct for transient). General rule for reviewers: for any new enum-like output value, grep for consumers that string-match or tokenize the value, and trace the new value through each consumer's logic explicitly — the fact that early exits write no run file made this a latent property rather than a live bug, but the same code reaching a persisted run error would have made it verdict-affecting.
