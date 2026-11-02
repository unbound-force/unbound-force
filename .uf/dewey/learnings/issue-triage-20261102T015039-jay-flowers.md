---
tag: issue-triage
author: jay-flowers
category: context
created_at: 2026-11-02T01:50:39Z
identity: issue-triage-20261102T015039-jay-flowers
tier: draft
---

Triage verification of refactor follow-up issue #741 (Decompose uf-workflow plugin index): every claim was objectively verifiable — grep confirmed parseTasksText/matchTaskLine/TASK_LINE_PATTERN (L3880-4131) and mostRecentArtifact/readDispatchArtifacts (L2879-2909) in .opencode/plugins/uf-workflow/index.ts, and the cited extraction pattern exists (.opencode/lib/ holds 12 modules including branch-classification.ts and review-dispatch-artifact.ts; note glob for `.opencode/lib/*.ts` returns nothing due to hidden-directory matching — read the directory instead). The file grows monotonically as tools are registered: 3,700+ lines at issue filing (2026-10-16) vs 4,300+ at triage (2026-11-01), so decomposition debt compounds until the refactor lands. Review-council follow-up issues that cite a dispatch correlation ID and name exact functions triage as objective; duplicate check by scope, not topic: #769 (review-dispatch test helpers) is related but non-duplicate.
