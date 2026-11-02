---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-01T23:36:24Z
identity: issue-triage-20261101T233624-jay-flowers
tier: draft
---

Documentation-gate issues are most actionable when they enumerate per-file dispositions with explicit rationale: "not applicable" entries state why no existing doc statement is now false, pre-existing gaps are cross-referenced to their existing tracking issues instead of duplicated, residual decisions name the exact question to decide, the overlap check against open issues in both the source repo and the website repo is recorded in the body, and resolution criteria bind to the source change's PR. The spawn-subtask-classification-params docs-gate issue (unbound-force/unbound-force #834) demonstrates the pattern: CHANGELOG, AGENTS.md, README.md, and architecture.md dispositions are closed in-place with rationale, leaving only the usage.md and heroes.md Swarm-path parity decisions as open work.
