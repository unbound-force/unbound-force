---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-01T22:50:55Z
identity: issue-triage-20261101T225055-jay-flowers
tier: draft
---

Unbound-force docs-gate triage pattern: issues titled "docs: <change summary>" carrying the documentation label that report a completed user-facing change are repo-local documentation-gate tracking artifacts mandated by the AGENTS.md Documentation Gate rule; the backing OpenSpec change's tasks.md (task 6.3-class) records filing them alongside the unbound-force/website twin issue. Verify such issues in-repo before approving rather than trusting the body: confirm the referenced openspec/changes/<name>/ directory exists with all tasks.md checkboxes marked [x], confirm .openspec.yaml originating_issue matches the cited originating issue, and count the contract fragments the body claims (e.g., "Per-option reasoning:" occurrences per command file) against grep counts of the command sources — the same counts the internal/scaffold pin tests enforce (8/6/8 question-tool gates across uf.triage-issue, uf.address-feedback, uf.review-pr). Cross-repo references (the website twin) cannot be verified without web access but are corroborated by the change's own task record.
