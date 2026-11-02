---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-01T22:39:19Z
identity: issue-triage-20261101T223919-jay-flowers
tier: draft
---

Documentation-gate tracker issues are objectively triageable when they enumerate one disposition per documentation artifact with exact, diff-verifiable counts and explicit no-change dispositions with rationale. Pattern observed in unbound-force/unbound-force#862 (uf-fix-ci docs impact): CHANGELOG/AGENTS.md/docs/QUICKSTART rows each name the exact edit ("PR Review Commands table row added", "commands count 32 → 33", "embedded assets 11 → 12"), README carries a reasoned no-change disposition ("no per-command reference table"), and content-coverage decisions (tutorial/blog) state their rationale. Reviewers can then verify completeness by walking the artifact list row-by-row against the implementing PR diff instead of judging an aspirational "update docs as needed".
