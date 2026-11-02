---
tag: review-dispatch
author: jay-flowers
category: pattern
created_at: 2026-11-02T21:23:52Z
identity: review-dispatch-20261102T212352-jay-flowers
tier: draft
---

Guard review of fix iterations (iteration 2+): before re-flagging an item left unfixed from a previous round, check the spec's FR text — it may be spec-sanctioned behavior adjudicated by design rather than a defect. In opsx/preflight-project-checks iteration 2, SKILL.md's "drop the last column outside soft-gate" prose initially looked like an unresolved determinism gap (mechanical step left in agent prose), but PF-FR-020 explicitly defines `rendered` as soft-gate-format rows with the consuming skill dropping the causality column outside soft-gate, and the design notes record the mode-aware-rendering refinement as deliberately deferred. Re-flagging it would have demanded churn against the change's own acceptance criteria. Procedure: verify each iteration-1 finding's fix against the updated diff, and for adjudicated items confirm the spec text sanctions the behavior before treating them as persistent findings.
