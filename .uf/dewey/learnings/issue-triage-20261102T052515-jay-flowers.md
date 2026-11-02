---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T05:25:15Z
identity: issue-triage-20261102T052515-jay-flowers
tier: draft
---

Triage pattern for design-discussion issues: an issue that honestly asks "is this actually a problem?" and considers status quo as a legitimate outcome still triages as VALID/enhancement/objective when four signals are present: (1) the core factual claim cites verifiable file:line evidence, (2) options are enumerated including doing nothing, (3) the governing principles on both sides of the trade-off are named, and (4) a recommendation is stated. Verified on unbound-force/unbound-force#558 (ecosystem provenance vocabulary): the divergent structs — core Envelope (internal/artifacts/artifacts.go, field set corroborated by specs/009-shared-data-model/spec User Story 1), Gaze Metadata (gaze/internal/taxonomy/types.go:255), Dewey provenance fields (dewey/store/embeddings.go:23) — are checkable in code, and Dewey semantic search against specs/009-shared-data-model and schemas/envelope/README corroborated the referenced envelope contract. The deliberative tone does not make the issue subjective: only the choice among enumerated options is judgment; the factual core is verifiable. Triage's role is to confirm the evidence and route the decision (cross-repo scope routes to the strategic spec tier at implementation time), not to pre-empt the design choice.
