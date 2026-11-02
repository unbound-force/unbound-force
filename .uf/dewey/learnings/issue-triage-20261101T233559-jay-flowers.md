---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-01T23:35:59Z
identity: issue-triage-20261101T233559-jay-flowers
tier: draft
---

Triage validation pattern for documentation-gap issues that report a bounded selection criterion missing from reference docs: (1) confirm the constraint exists in the enforcing artifacts — the implementing command, skill, and deterministic tool guard (for tier-routing risk floors: tier-routing SKILL.md Phase 4 plus the uf_unleash_route_decision ROUTE_RISK_FLOOR guard carrying risk_signals); (2) confirm each named reference doc omits it — AGENTS.md tables verify directly from standard triage context, while docs/ page claims corroborate via issue provenance (audit PR, validated compiled article) when the page sits outside context; (3) classify as a documentation bug, not an enhancement, when the omission hides a tool-enforced MUST NOT — the doc misstates eligibility criteria for a class of changes rather than merely lacking polish. Validated on unbound-force/unbound-force issue #826: risk floors are documented in uf.unleash.md and the tier-routing SKILL.md (corroborated by the validated compiled/spec-workflow and compiled/uf-unleash articles) but absent from the AGENTS.md Quick row (verified in provided repo context) and the docs/usage.md workflow table (audit provenance: PR #820, documentation-pipeline compiled article "Still-open gap").
