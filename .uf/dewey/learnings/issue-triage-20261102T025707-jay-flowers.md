---
tag: issue-triage
author: jay-flowers
category: gotcha
created_at: 2026-11-02T02:57:07Z
identity: issue-triage-20261102T025707-jay-flowers
tier: draft
---

unbound-force AGENTS.md carries two identically-named "Documentation gate" bullets with different satisfaction conditions: (1) an impact-assessment gate (CHANGELOG/AGENTS/README updates) that retrospective Dewey learnings explicitly satisfy, and (2) a filing gate that MUST file a documentation issue in the current repo for user-facing changes before PR merge. Learnings satisfy only the first. Verified while triaging docs-gate issue #690 (filed for change #687 / PR #691): the PR's documentation footprint (CHANGELOG Fixed entry, openspec/changes/fix-stale-base-ref artifacts, 4 dewey learnings) satisfied the assessment gate, while #690 remains the open tracker for the actual user-facing docs/ updates (resolve_base_ref fallback behavior, review-council base-ref handling, Key Files Changed removal). When triaging documentation issues, do not close a docs-gate issue as satisfied by shipped learnings or a CHANGELOG entry — those satisfy the assessment gate, not the documentation work the issue tracks.
