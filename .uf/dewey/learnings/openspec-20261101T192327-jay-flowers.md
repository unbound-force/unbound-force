---
tag: openspec
author: jay-flowers
category: pattern
created_at: 2026-11-01T19:23:27Z
identity: openspec-20261101T192327-jay-flowers
tier: draft
---

Scenario-name continuity in an OpenSpec ADDED requirement has no mechanical force at archive: the scenario-continuity check compares a MODIFIED block against the current canonical spec, so carrying a predecessor delta's scenario names in an ADDED entry is commemorative only — doubly so when the predecessor's own delta is an unarchivable orphan whose scenario names can never reach the canonical spec. Verified on accept-variantless-parent-model (quick/accept-variantless-parent-model): the ADDED 'Finalization Model Validation' carries the null-model rejection under two names ('Host-resolved advisor without resolved parent fails' from advisor-host-model-fallback's orphan MODIFIED and 'Host-resolved advisor without resolved parent model fails' from the fix loop) with equivalent bodies, justified as lineage continuity. Prefer one canonical scenario name per behavior in ADDED entries; record lineage in design.md prose, which already carries the naming history.
