---
tag: issue-triage
author: jay-flowers
category: gotcha
created_at: 2026-11-02T05:25:20Z
identity: issue-triage-20261102T052520-jay-flowers
tier: draft
---

Triage re-verification gotcha from unbound-force issue #558: re-verify an issue's cited consumers against the current tree at triage time, because the ecosystem may have contracted after filing. #558 motivated provenance-vocabulary unification with "Mx F metrics collection already reads envelopes from Divisor, Muti-Mind, and Gaze via artifacts.ReadEnvelope", but the remove-mutimind-mxf-orchestration OpenSpec change (issues #565/#566, all tasks complete) later deleted Mx F, Muti-Mind, cmd/mxf/, cmd/mutimind/, internal/metrics/, and the orchestration engine — cmd/ now holds only coverage-gate and unbound-force — and VISION.md, the issue's discovery context, no longer exists in the repo. The divergence observation itself stayed verifiable (internal/artifacts/artifacts.go Envelope vs Gaze's own report schema vs Dewey's search-result provenance fields), so the issue remained a valid enhancement with stale motivation rather than invalid. Pattern: verify every named consumer, binary, package, and document in an issue's motivation against the current tree, and check for higher-numbered issues that may have removed the cited components, before weighting the issue's motivation.
