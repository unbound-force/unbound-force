---
tag: spec-review
author: jay-flowers
category: gotcha
created_at: 2026-11-02T01:35:39Z
identity: spec-review-20261102T013539-jay-flowers
tier: draft
---

OpenSpec spec-review iteration fixes must propagate to every sibling artifact that enumerates the same information. Observed in init-override-stubs-gitignore-sync iteration 2 (2026-11-01): the iteration-1 fix that added task 6.4 (docs/architecture.md Managed-files refresh) was applied to tasks.md only; proposal.md's What Changes item 10 and Impact Modified list still omitted docs/architecture.md, resurfacing as a MEDIUM cross-artifact documentation-impact inconsistency in the next review iteration. When a review fix adds, removes, or re-scopes a task, sweep the sibling artifacts — proposal What Changes/Impact file lists, design Decisions/Goals, spec delta numbering headers — for enumerations covering the same files and requirements, and update them in the same fix.
