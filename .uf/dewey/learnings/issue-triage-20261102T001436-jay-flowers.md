---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T00:14:36Z
identity: issue-triage-20261102T001436-jay-flowers
tier: draft
---

A documentation-impact tracking issue in the unbound-force org is complete when it dispositions every artifact named in the AGENTS.md documentation gate — CHANGELOG.md, AGENTS.md, README.md, plus repo user-facing docs such as docs/usage.md and docs/architecture.md — records the cross-repo website tracking issue number, and states a content-coverage decision (blog/tutorial) with a significance rationale. Explicit "no change needed" entries carrying justifications (e.g., "AGENTS.md documents neither the finale step sequence nor per-command flow") are as important as "updated" entries: silence on a gate-named artifact is an incomplete disposition record and is grounds for NEEDS-CLARIFICATION during triage. Issue #810 (rebase-upstream-before-push docs impact) is the reference exemplar: every gate artifact received an explicit disposition, the website cross-reference (unbound-force/website#320) was verifiable via Dewey, and no duplicate tracking issue existed.
