---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-01T23:38:41Z
identity: issue-triage-20261101T233841-jay-flowers
tier: draft
---

Documentation-gate issues in unbound-force repos (filed per the AGENTS.md "MUST file a documentation issue before PR merge" rule) triage quickly and reliably via three anchor checks, each a single glob/grep: (1) the referenced change directory exists (openspec/changes/<name>/ or specs/NNN-*/), (2) every enumerated page exists, and (3) the described behavior matches the implementation while the claimed stale wording matches the current docs. Validated on issue #827 (fix-init-speckit-sentinel docs): the change directory existed with all artifacts, all four target pages existed, the `.specify/scripts` sentinel matched internal/scaffold/scaffold_test.go, the stale "silently skipped" wording sat at docs/architecture.md:308, and the dewey-only `method: skip` note sat at docs/cli-reference.md:39 — five checks, all green, VALID in minutes. When any anchor fails (change missing, page absent, behavior not in code, stale wording not found), route to NEEDS-CLARIFICATION rather than assuming the issue runs ahead of an unmerged branch.
