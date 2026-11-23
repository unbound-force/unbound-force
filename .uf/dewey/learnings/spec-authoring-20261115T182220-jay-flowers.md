---
tag: spec-authoring
author: jay-flowers
category: gotcha
created_at: 2026-11-15T18:22:20Z
identity: spec-authoring-20261115T182220-jay-flowers
tier: draft
---

A tool-extraction spec that mandates validation against a target schema must state the derivation rule for every REQUIRED field of that schema, not only the fields the old prose happened to describe. In the feedback-triage-artifact-tool spec, the mapping rules covered exactly the two summary fields the 1.1.0 schema leaves optional (filed/replied) while all eight required summary fields (total_items, accepted, modified, rejected, asked, tier1_count, tier2_count, divisor_agents_involved) had no stated derivation — a tool built exactly to the spec's mapping rules fails its own mandated validation. Check the schema's required list against the spec's mapping rules before approving an extraction spec.
