---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-01T21:54:25Z
identity: issue-triage-20261101T215425-jay-flowers
tier: draft
---

When triaging documentation follow-up issues in unbound-force repos, recognize them as documentation-gate artifacts rather than redundant trackers: AGENTS.md requires filing a documentation issue against the current repo for user-facing changes before PR merge, and the constitution's Cross-Repo Documentation rule requires a companion website-repo issue. Such issues triage cleanly as VALID/enhancement (objective) when they (1) explicitly scope what the implementing PR's doc updates already cover versus remaining follow-up work, (2) cite the originating issue, the spec/openspec change directory, and the cross-repo website tracking issue, and (3) carry concrete checklist items. Conditional release-time checklist items (e.g., "next release notes should state X") are acceptable low-cost reminders tied to real trade-offs, not zero-waste violations. Corroborate claims against the openspec change's proposal Impact section via Dewey before ruling — the spec artifacts enumerate the affected doc surfaces and completion state.
