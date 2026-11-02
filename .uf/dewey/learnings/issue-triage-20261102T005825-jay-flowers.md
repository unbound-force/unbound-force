---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T00:58:25Z
identity: issue-triage-20261102T005825-jay-flowers
tier: draft
---

Documentation-gate tracker issues for OpenSpec changes in unbound-force/unbound-force are high-confidence VALID triage targets when they carry four verifiable signals: per-file dispositions for every gate-named surface (CHANGELOG, usage.md, heroes.md, architecture.md, AGENTS.md, README.md) traced to the change's Verification Record, a blog/tutorial reassessment performed at filing, a documented overlap check, and a close condition tied to the originating PR. Each claim is cheaply corroborable against the repo — grep the CHANGELOG entry, the agent-checklist section, and the docs for contract identifiers (e.g., task_classification / SKILL_REFERENCE_REQUIRED). When keyword searches surface structurally similar docs-gate issues (e.g., #834 spawn-subtask-classification-params vs #791 delegation-skill-enforcement), differentiate by originating change rather than by shared governance rule or touched files: the follow-up change MODIFIES the same DE-008 requirement and also extends the divisor-guard checklist, so both docs issues are legitimate siblings whose heroes.md dispositions still require cross-issue coordination to avoid conflicting edits.
