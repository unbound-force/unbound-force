---
tag: spec-governance
author: jay-flowers
category: gotcha
created_at: 2026-11-01T17:49:42Z
identity: spec-governance-20261101T174942-jay-flowers
tier: draft
---

Reshaping an OpenSpec requirement from MODIFIED to ADDED (the orphan-rescue move when no predecessor text can ever exist) collides two MUST rules: MP-002/MP-013 require every requirement to carry a stable FR-NNN id line, while `openspec validate --strict` enforces a ~500-character description budget on ADDED requirements only (MODIFIED is exempt so existing text stays whole). The id line consumes that budget, so the reshape tempts dropping the identifier to stay under the limit — leaving a new requirement uncitable by FR id and internally inconsistent with id-carrying siblings in the same delta (observed on accept-variantless-parent-model, quick/accept-variantless-parent-model branch, 2026-11-01: the ADDED 'Finalization Model Validation' carries no id while the same delta's MODIFIED 'Review Dispatch Artifact' opens with '**ID: RD-FR-008**'). Plan the FR allocation before the reshape and re-run strict validation after adding the id line; if the budget forces a choice, record the id omission as an explicit convention-pack exception in design.md rather than leaving it implicit.
