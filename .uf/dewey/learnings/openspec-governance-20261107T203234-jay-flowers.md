---
tag: openspec-governance
author: jay-flowers
category: pattern
created_at: 2026-11-07T20:32:34Z
identity: openspec-governance-20261107T203234-jay-flowers
tier: draft
---

Bounding a retroactive pin-supersession sweep after a dependency security bump: grep the unarchived openspec/changes/ stack for the moved values (old version string and old baseline hash), then classify hits into three remedy tiers. (1) MUST-text requirement bodies are the blocking tier — e.g. clarification-marker-superset SC-FR-003 pinning zod 4.1.8 and rollback SHA 5ebee4af... against shipped 4.6.6/1bdb0793... (PR 948, feedback item afb-948-item2-design-md-11); they need a SUPERSEDED-BY annotation on the latest-generation carrier, because MODIFIED restatement replaces the text wholesale at archive. (2) Requirement text already carrying a SUPERSEDED-BY annotation whose "remainder ... remains in force" clause still holds the old value (review-council-multi-model-fanout spec L64) is covered once the successor restatement carries the annotation — annotating the predecessor too is optional cheap belt-and-braces. (3) Design/tasks provenance records and rationale prose naming the old version ("zod 4.1.8 is already a runtime dependency" in fix-tool-args-schema-mismatch) are historical context, not pin mandates — exclude them explicitly in the change's Non-Goals rather than annotating, so the sweep-scope enumeration stays accurate. Verification discipline for the annotation text: cite the CHANGELOG entry verbatim, because the triggering review thread's vulnerability citations can be truncated or garbled (the thread's "GHSA-443g9" did not match CHANGELOG's GHSA-mrhx-c7xc-m3w5, and toml carried 4 GHSAs, not the thread's 3).
