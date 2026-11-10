---
tag: documentation
author: jay-flowers
category: pattern
created_at: 2026-11-08T23:16:46Z
identity: documentation-20261108T231646-jay-flowers
tier: draft
---

unbound-force CHANGELOG entries close with a parenthesized traceability footer — "(Spec: openspec/changes/<name>/)" and, for user-facing changes, "Docs: #<N>" referencing the pre-merge documentation issue. House examples: CHANGELOG.md "(Spec: openspec/changes/stop-shipping-plugin-lock/, Docs: #964, unbound-force/website#382)" and "(Spec: openspec/changes/determinism-codification/)". New entries that omit the Spec: path fail the AGENTS.md documentation gate even when the prose is complete, and the footer is also the cheapest place to surface the constitutionally required documentation issue for user-facing behavior changes.
