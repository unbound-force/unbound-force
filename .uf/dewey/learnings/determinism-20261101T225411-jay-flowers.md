---
tag: determinism
author: jay-flowers
category: context
created_at: 2026-11-01T22:54:11Z
identity: determinism-20261101T225411-jay-flowers
tier: draft
---

The rewritten /uf.run-pipeline-plan command (update-run-pipeline-plan change, PR #839) deliberately retains four orchestrator-side mechanical steps as command prose — paused-exit heading recognition, the [BLOCKED: <variant>] todowrite annotation, the disk-handoff verification loop, and the closes-linkage gh pr view / gh pr edit backstop — as a documented DE-001 trade-off recorded in the change proposal's Principle VI section under the determinism pack's Conflict Resolution rule. Issue #842 tracks their extraction into tested uf-workflow plugin tools. Reviewers who encounter mechanical prose in this command should cite #842 as the tracked deferral instead of filing a new determinism violation. Implementation note: the paused-exit variant-to-re-run-class mapping (mid-pipeline vs tier-routing variants) is deterministic routing that belongs inside the recognition tool per DE-006, not re-derived by the orchestrator.
