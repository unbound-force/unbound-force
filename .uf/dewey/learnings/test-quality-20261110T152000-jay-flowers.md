---
tag: test-quality
author: jay-flowers
category: gotcha
created_at: 2026-11-10T15:20:00Z
identity: test-quality-20261110T152000-jay-flowers
tier: draft
---

When a spec enumerates a decision set (e.g., FRT-FR-001: "resolve the threads of accept, modify, file, and reply review items"), representative-scenario testing pins only the values that happen to appear in the chosen scenarios. Excluded-but-plausible values (the "ask" decision — a plausible member of the same enum) and cross-product cells (reply decision × unresolved thread) stay unpinned, so a set-membership regression in the predicate (RESOLVABLE_DECISIONS gaining or losing a member) ships silently even under a 94% statement-coverage floor, because the uncovered branches are data-dependent, not statement-dependent. Rule: when a requirement enumerates a finite set, add a table-driven test over the full set including the excluded values, asserting both the positive members and the exclusions.
