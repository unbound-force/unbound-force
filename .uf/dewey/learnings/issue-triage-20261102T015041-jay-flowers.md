---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T01:50:41Z
identity: issue-triage-20261102T015041-jay-flowers
tier: draft
---

Keyword-based duplicate searches over issue titles miss scope-level overlaps between refactor issues. When triaging extraction/refactor issues in unbound-force, supplement the provided keyword duplicate candidates with a semantic knowledge-graph search (Dewey) for overlapping scopes: issue #741 (decompose uf-workflow/index.ts: task-line grammar, git porcelain helpers, checkpoint state store) appeared unique under keyword search (nearest candidate #769, review-dispatch test-helper consolidation), but semantic retrieval surfaced #782 (extract issue-reference helpers + shared progress-state path/guard helpers from the same plugin suite), whose state-path scope overlaps #741's checkpoint-state-store extraction. Neither issue subsumes the other, so the correct disposition is cross-referencing and sequencing rather than duplicate closure — and only the semantic check revealed the overlap.
