---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T02:37:25Z
identity: issue-triage-20261102T023725-jay-flowers
tier: draft
---

Documentation-tracking issues that reference an OpenSpec change can be triaged objectively without network access by checking three repo anchors: (1) the referenced openspec/changes/<name>/ spec delta — confirm the FR range exists and its requirements match the issue's behavioral claims; (2) the affected .opencode/commands/ files — grep for the claimed behavior being present (e.g., "consolidate_dispatch exactly once") and superseded patterns being absent (e.g., --slurp, direct finalize_review_dispatch calls); (3) the MCP tool schemas in the review environment — consolidate_dispatch's dual-mode provenance contract and the kind-native workflow_result enums (triage: VALID/INVALID/NEEDS-CLARIFICATION, feedback: ACCEPT/AUTHOR-DECIDES) are directly readable from tool definitions, so tool-behavior claims in issues are checkable without reading plugin source. A final grep of docs/ for the tool name confirms whether the documentation gap is real (issue #717: consolidate_dispatch appears only in learnings review files, never in user-facing docs/usage.md or docs/architecture.md). Issue #717 passed all checks: FR-001 through FR-008 matched, all three commands carry the one-call mandate and the finalize_review_dispatch prohibition, and the docs gap is confirmed.
