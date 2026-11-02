---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-01T22:20:45Z
identity: issue-triage-20261101T222045-jay-flowers
tier: draft
---

Documentation-gate issues in unbound-force repos (filed per the AGENTS.md rule "MUST file a documentation issue against the current repo for user-facing changes before PR merge") triage efficiently as VALID/enhancement/objective using four checks: (1) the cited gate text exists in the repo's AGENTS.md; (2) the originating change ships user-facing surfaces outside the exempt categories — config fields, tool input/output contracts, and command prose all qualify; (3) the issue precedes the feature PR merge; (4) duplicate search returns none. Two verification greps catch both false directions: a repo-wide grep of the surface token (e.g. per_run_timeout) confirms the implementation exists, while a targeted grep of AGENTS.md confirms the documentation gap is unfilled. Strongest grounding signal: the change's own tasks.md contains the documentation-gate task (derive-per-run-timeout task 8.3 mandated the in-repo docs issue), making the issue a planned deliverable — and the CHANGELOG entry corroborates by recording the user-facing classification and the filed gate issues.
