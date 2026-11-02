---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T02:58:35Z
identity: issue-triage-20261102T025835-jay-flowers
tier: draft
---

Docs-gap issue triage becomes verification rather than plausibility-checking when each enumerated claim is cross-referenced against the indexed spec/design/tasks artifacts of the cited source change. Applied to unbound-force issue #676 (document tier_caps in .uf/review-matrix.yaml): every enumerated item matched the tier-gated-agent-count requirements — the limits.tier_caps field (FR-003), defaults lightweight:2/standard:null/heavy:null (FR-003, D3), floor enforcement max(cap, always_required_count) (FR-001), the --full guard (task 2.2), the priority order always-required > scope-match alphabetical > curator-relevant (FR-001), and the HIGH-severity heavy-cap advisory on security-sensitive diffs (FR-006, D7, Scenario 12). The cross-check also surfaced an artifact conflict the issue text alone could not reveal: tasks.md task 2.3 specifies MEDIUM for that same advisory, contradicting FR-006's HIGH mandate — the docs author must verify the implemented severity before writing, since documentation records what the code emits, not what any single artifact claims. Cohesion rule: enumerated items that all bind to one field from one change signal completeness, not scope creep — keep split_recommendation null.
