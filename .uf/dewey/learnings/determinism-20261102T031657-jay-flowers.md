---
tag: determinism
author: jay-flowers
category: pattern
created_at: 2026-11-02T03:16:57Z
identity: determinism-20261102T031657-jay-flowers
tier: draft
---

Standalone lookup-tool proposals that serve a single deterministic consumer conflict with the Router Pattern (determinism pack DE-005): a judgment-free call-then-use sequence — call the lookup tool, read its structured result, pass it to the next call — is mechanical orchestration that MUST collapse into the consuming tool, with the lookup living inside as a constant or pure function (DE-006 names verdict mapping explicitly as such a routing decision). Triage disposition for these issues: keep the extraction issue open only if a consumer for the lookup survives outside the collapsed path; if it proceeds, require the mapping constants to live in the shared library (e.g., .opencode/lib/) alongside existing projections (such as consolidate_dispatch's canonicalDecision) so a single source of truth is maintained instead of a second divergent table. Observed triaging unbound-force/unbound-force#671 (map_verdict) against #664 (post_github_review internal constant).
