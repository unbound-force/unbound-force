---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T01:13:51Z
identity: issue-triage-20261102T011351-jay-flowers
tier: draft
---

Documentation-tracking issues for merged behavior changes are far easier to triage when they explicitly delimit scope against adjacent issues. Issue #778 (docs for the fix-dispatch-verdict-derivation change) names its antecedent (#717, per-run verdict layer) and the adjacent non-overlapping docs issues (#766 recovery tools, #754 data-flow corrections), so duplicate-scoping reduces to checking the named boundaries instead of re-deriving scope from keyword search. The issue also maps each checklist item to a verifiable artifact (spec FR numbers, command-contract section numbers, scaffold-mirror paths), letting the triager confirm validity with targeted file reads rather than a full code review. Verified during triage: spec delta FR-009..FR-012, uf.address-feedback.md section 2.3.5, and the scaffold mirror path all existed exactly as cited, while docs/ had zero coverage of the new verdict semantics — confirming the documentation gap was real.
