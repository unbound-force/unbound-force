---
tag: instruction-quality
author: jay-flowers
category: gotcha
created_at: 2026-11-28T17:45:11Z
identity: instruction-quality-20261128T174511-jay-flowers
tier: draft
as_of: "unbound-force@444cb92a90a02f80a8798f4de1152ff133d7abe4"
---

An aggregate job with `if: always()` that serves as the sole required check enforces through its result-verification step, not through its `needs` array: `needs` only makes job results visible, while the verification step (env vars mapping `needs.<job>.result` plus a closed `job=result` pair loop) is what fails the check. Instruction prose that tells agents how to wire a new gated job into such an aggregate must name every edit — the `needs` entry, the result env var, and the loop pair — or explicitly point at the verification step as the checklist. A needs-only sentence produces a silent gap: the new job can fail while the required check stays green.
