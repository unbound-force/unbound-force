---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T05:25:33Z
identity: issue-triage-20261102T052533-jay-flowers
tier: draft
---

Split signal for bundled multi-gap issues: fix-dependency asymmetry, not shared theme. unbound-force/unbound-force#561 bundled two input-validation gaps (ReadEnvelope schema validation in internal/artifacts; ProjectDir traversal validation in internal/gateway) sharing only an audit-origin theme — different packages, different root causes, no causal link. Gap 1's fix was blocked on #557 (the envelope schema lists `context` as a required field — corroborated by schemas/envelope README and specs/009 — but the writer emits it omitempty, so enforcing validation first would reject existing artifacts), while Gap 2 was immediately actionable by reusing an existing validated pattern (metrics/security.go ValidateProjectPath). When one bundled part is dependency-blocked and another is actionable, split: a shared theme is not coupling, and bundling entangles an actionable fix with a blocked one. Constitution Principle V (file contents and user input MUST be validated before security-sensitive operations) is the objective yardstick that classifies missing-validation reports as bugs — MUST-rule compliance defects — rather than enhancements.
