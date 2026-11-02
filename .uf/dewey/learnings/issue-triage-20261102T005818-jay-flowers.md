---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T00:58:18Z
identity: issue-triage-20261102T005818-jay-flowers
tier: draft
---

Documentation-gate issues filed per change in unbound-force repos arrive in structurally parallel shapes (CHANGELOG disposition, per-file docs dispositions, blog/tutorial reassessment, overlap check), so two open docs-gate issues with near-identical bodies are usually parallel gate records for different changes, not duplicates. When triaging, verify the underlying change identity — the openspec/changes/<name>/ directory and originating issue referenced in the body — before classifying as duplicate. Observed with #791 (delegation-skill-enforcement) and #834 (spawn-subtask-classification-params): identical structure, different changes.
