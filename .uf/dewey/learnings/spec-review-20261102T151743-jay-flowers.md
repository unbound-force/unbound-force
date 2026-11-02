---
tag: spec-review
author: jay-flowers
category: pattern
created_at: 2026-11-02T15:17:43Z
identity: spec-review-20261102T151743-jay-flowers
tier: draft
---

When reviewing an OpenSpec delta whose MODIFIED entries supersede a predecessor change's text, verify two chains before approving. First, the MP-014 numbering chain: the predecessor's 'reserved for future' range must be consumed exactly by the successor's allocation — host-fallback-provenance reserved RD-FR-038..048, and split-host-resolution-failure-classes allocates RD-FR-038 while re-reserving 039..048 — and a repo-wide grep for the allocated id must show no competing claim (verified clean for RD-FR-038 on 2026-11-02). Second, predecessor-text fidelity: read the predecessor delta and diff the restated requirement scenario-by-scenario to confirm the MODIFIED entry changes only the clauses the change intends (matching the header's 'narrows only X' claim), carries every other scenario forward verbatim, and ends with an accurate Previously paragraph. On split-host-resolution-failure-classes both chains held: all nine RD-FR-037 scenarios carried forward with only the failure clause narrowed, the Advisor Host-Model Fallback requirement inherited its unnumbered shape faithfully from advisor-host-model-fallback through host-fallback-provenance, and both Previously paragraphs accurately described the superseded contracts.
