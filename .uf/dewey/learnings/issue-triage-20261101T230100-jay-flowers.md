---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-01T23:01:00Z
identity: issue-triage-20261101T230100-jay-flowers
tier: draft
---

Docs-gate issues that summarize system behavior — not just surface deltas — carry a distinct triage risk: their behavioral claims can be falsified by observation. Issue #838 (update-run-pipeline-plan) shows both the failure and the repair: the original body asserted gates "surface directly to the human from the delegated child sessions," falsified by observed delegation timeouts at question gates (parented child sessions cannot surface question-tool prompts to the human); the corrected body carries dated provenance (2026-10-27, PR #839 review feedback, learning uf-workflow-20261026T182050). Triage pattern: corroborate behavioral summaries in docs-gate issues against the originating change's authored design record — the update-run-pipeline-plan design doc's Risks/Trade-offs independently documents the same wedge-at-timeout behavior and the gh pr edit backstop — and treat dated correction provenance as a confidence signal rather than a defect. Companion pattern: when the triage pipeline's duplicate search returns none but the issue body self-reports a found duplicate, an explicit disposition resolves the question without invalidating either issue — #838 supersedes #747's behavioral assessment (three-subagent pipeline, gate relay) while deliberately keeping #747 open for its distinct AGENTS.md follow-up, so the two issues track different scopes and neither is a duplicate.
