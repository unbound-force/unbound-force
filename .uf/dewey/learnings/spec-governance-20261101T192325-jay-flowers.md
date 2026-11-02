---
tag: spec-governance
author: jay-flowers
category: gotcha
created_at: 2026-11-01T19:23:25Z
identity: spec-governance-20261101T192325-jay-flowers
tier: draft
---

A retrospective learning and the artifact it diagnoses can ship in the same PR without the artifact being repaired: accept-variantless-parent-model stored learning openspec-20261101T174938 (prescribing "grep the capability's highest allocated id across openspec/changes/*/specs/<capability>/spec.md, allocate the next one, and re-run openspec validate --strict" before authoring an ADDED requirement) while its own ADDED 'Finalization Model Validation' shipped id-less, and design.md's non-goal defended the omission with the MP-002 no-renumbering argument the learning explicitly refutes (MP-002 governs existing identifiers, not new allocations; the id line fits the 500-char ADDED budget per RD-FR-028/RD-FR-032 precedent). Guard: when a fix-loop retrospective names a concrete, bounded repair to an artifact in the current change, convert it into a tasks.md item in the same iteration rather than knowledge-capture only — the cheapest fix window closes at merge. Here, archive would make the id-less requirement permanent in the canonical spec, and repairing it later costs an entire additional OpenSpec change.
