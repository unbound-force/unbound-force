---
tag: review-dispatch
author: jay-flowers
category: gotcha
created_at: 2026-11-02T03:54:23Z
identity: review-dispatch-20261102T035423-jay-flowers
tier: draft
---

In review-dispatch (.opencode/plugins/review-dispatch/index.ts), the change-profile `security_sensitive` flag has three consumers, not one: `diffTier` forces the heavy review tier, `shouldAugment` auto-adds divisor-adversary as a reviewer, and dispatch planning emits a HIGH advisory when tier_caps.heavy caps a security-sensitive change. An issue or fix targeting only the visible consumer (tier forcing) can silently change reviewer composition: narrowing the flag's path heuristics (e.g., the over-broad `/(^|\/)(cmd|internal)(\/|$)/` segment match from issue #656) also removes the Adversary from changes in packages like internal/gateway and internal/sandbox whose file names do not match the `/(secret|token|key)/` keyword regex. Enumerate every consumer of a classification flag before changing its inputs — over-inclusive classification is a cost bug, but over-narrowed classification is a security bug.
