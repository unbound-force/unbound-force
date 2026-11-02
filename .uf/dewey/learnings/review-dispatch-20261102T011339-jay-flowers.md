---
tag: review-dispatch
author: jay-flowers
category: context
created_at: 2026-11-02T01:13:39Z
identity: review-dispatch-20261102T011339-jay-flowers
tier: draft
---

In the unbound-force review-dispatch plugin, the top-level workflow verdict (workflow_result.value and verdict) is vote-derived via calculatedWorkflowValue — the same function the finalizer validates with — while verdict_reason and advisories remain severity-derived (computeVerdict) human context. Sharing one derivation function between emission and validation makes emission/validation disagreement structurally impossible; the prior bug (#764, observed on PR #761) arose from a severity ladder at emission diverging from vote-based validation on split panels. Split-panel outcomes per the fix (spec FR-009..FR-012 in openspec/changes/fix-dispatch-verdict-derivation/): feedback panels with any successful zero-findings ACCEPT run finalize ACCEPT/APPROVE per the strictest rule (uf.address-feedback.md §2.3.5); triage split-majority panels finalize as the persona majority; zero-successful-runs panels finalize with the no-success pass-through verdict (INCONCLUSIVE/UNAVAILABLE). Council and test-review outputs are unchanged because both derivations already agree on every input there.
