---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T01:32:20Z
identity: issue-triage-20261102T013220-jay-flowers
tier: draft
---

Triage of Router Pattern (intent-level tool collapse) proposals maps directly onto the determinism pack's rule ladder: DE-005 mandates collapsing judgment-free tool sequences into single intent-level tools; DE-006 requires the proposal to state its judgment boundary (what stays with the LLM); DE-007 requires the implementing tasks.md to carry seam-verification tasks naming the three layers (registration surface with input-schema parity, documented output shape, write→read round-trips) — codified from issue #765 after the #716 migration stream broke at its seams in live use (#732: the @opencode-ai/plugin tool() helper silently ignores non-zod args shapes); and DE-004 operationalizes qualitative size claims like "materially fewer steps" as measurable line/token deltas. Hallmarks of a well-formed collapse proposal: constraints verified against the platform (SDK client exposes no tool-invocation method; per-run timeout ceiling bounds tool granularity), phased decomposition with per-phase spec tiers (OpenSpec opsx/<name> per phase), and acceptance criteria whose done-state requires only the feasible phase to ship plus a documented spike for the gated phase.
