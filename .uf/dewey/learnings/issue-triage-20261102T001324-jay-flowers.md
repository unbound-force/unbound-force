---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T00:13:24Z
identity: issue-triage-20261102T001324-jay-flowers
tier: draft
---

Duplicate-search collisions among documentation-gate issues: when two behavior changes touch the same user-facing surface, both resulting docs issues list the same affected files and keyword-collide (example pair: unbound-force#819, docs for the unleash-build-test-fix-loop change / PR #815, vs #840, docs for the interactive clarify ask gate — both cite .opencode/commands/uf.unleash.md and its embedded mirror). Disambiguate by change linkage — the PR number or openspec change name each issue tracks — not by overlapping file paths or issue phrasing; same-surface docs issues from different changes are separate valid issues, and only same-change issues are duplicates. Corroborate a gate issue's mandate via the linked change's tasks.md, whose documentation-gate task records the filed issue numbers (unleash-build-test-fix-loop task 3.2 records both website#322 and repo companion #819).
