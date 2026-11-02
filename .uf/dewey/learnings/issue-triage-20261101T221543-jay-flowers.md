---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-01T22:15:43Z
identity: issue-triage-20261101T221543-jay-flowers
tier: draft
---

Duplicate-check pattern for workflow-tutorial issues: distinguish the deliverable type before matching candidates. A command's documentation-gate issue (filed when the command lands, per the AGENTS.md documentation gate — e.g., unbound-force/unbound-force#862 "docs: /uf.fix-ci new command") tracks reference documentation and closes when the usage/reference pages exist; a workflow-tutorial issue (e.g., #870 for /uf.fix-ci, precedent #855 for /uf.rebase-pr) tracks a narrative walkthrough of the command's phases, gates, and budgets for practitioners. Both reference the same command but are complementary deliverables: the docs issue completing (docs/usage.md section merged) does not satisfy the tutorial's significance threshold (multi-phase workflow, human-decision gates, taxonomy, iteration budget), and a tutorial issue is a duplicate only of another tutorial issue for the same command. Verified during triage of unbound-force issue #870 against candidates #862 (docs-gate issue, work complete) and #873 (plugin-tool implementation issue) — neither is a tutorial deliverable.
