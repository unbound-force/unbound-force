---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T05:25:11Z
identity: issue-triage-20261102T052511-jay-flowers
tier: draft
---

Triage pattern for Unbound Force ecosystem-standardization issues: divergent provenance vocabularies across tools (uf Envelope in internal/artifacts/artifacts.go, Gaze Metadata in gaze/internal/taxonomy/types.go, Dewey search-result provenance fields) are a documented, deferred SHOULD-level gap — Spec 002 hero-interface-contract research explicitly recorded Gaze's non-envelope report output as "a known gap, not a blocker" with adoption deferred until Spec 009 formalized the schemas. Constitution Principle I's Envelope MUST covers inter-hero interchange artifacts, not a hero's primary standalone output; Principle III mandates provenance content (producer, version, timestamp, input context) without mandating field names, and registered-schema conformance is SHOULD-level. Classify such divergence issues as enhancement, not bug. When a shared provenance header is proposed, verify the field set covers all Principle III requirements (including input context: branch/commit) and reuses existing Envelope field names (schema_version) rather than minting synonyms (format_version).
