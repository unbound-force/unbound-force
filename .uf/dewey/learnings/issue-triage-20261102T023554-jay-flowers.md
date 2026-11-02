---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T02:35:54Z
identity: issue-triage-20261102T023554-jay-flowers
tier: draft
---

Documentation-issue triage pattern (verified on unbound-force #706, checkpoint/get_progress/clear_progress docs): a docs issue filed at feature-merge time enumerates consumers and doc targets as of that date, and both go stale silently. At triage time, verify three things against the tree instead of trusting the issue body. (1) Confirm the gap is real by grepping the user-facing surfaces (docs/, README.md) for the feature's identifiers — #706's gap is confirmed: zero matches for get_progress/clear_progress/.uf/progress across all 11 docs/ files and README.md, leaving a CHANGELOG entry as the only user-visible trace. (2) Re-derive the consumer list by grepping the implementation — the issue lists four command files, but uf.fix-ci.md also calls clear_progress (adopted after the issue was filed), and CHANGELOG records six files touched; docs issues inherit their scope snapshot from filing date, not from the current tree. (3) Distinguish doc targets from historical records — an ask like "update any documentation that references the old pattern" can have zero live targets when the old pattern was never user-documented: execution-checklist references exist only in immutable openspec/changes/ records, dewey learnings, and CHANGELOG history, which must not be rewritten. When (3) fires, the real deliverable is net-new documentation, not reference updates.
