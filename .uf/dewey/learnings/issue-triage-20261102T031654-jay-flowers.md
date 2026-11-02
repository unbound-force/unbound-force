---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T03:16:54Z
identity: issue-triage-20261102T031654-jay-flowers
tier: draft
---

Triage heuristic for deterministic-lookup extraction issues: when an issue proposes a standalone agent-callable lookup tool (for example, mapping review-council verdicts to GitHub review events), evaluate it against the Router Pattern rules (determinism pack DE-005/DE-006) before accepting the standalone shape. When the only consumer is a single downstream intent-level tool (for example, a shared PR-review posting tool), the mapping belongs as an internal constant of that tool: a standalone lookup preserves the agent-orchestration hop (call lookup, interpret result, then act) that the Router Pattern removes, while delivering the same token savings. Keep the standalone-tool issue open only when a second consumer of the mapping survives outside the primary path; otherwise disposition it as superseded when the consuming tool lands. Surviving mapping constants belong in the shared plugin library (.opencode/lib/) so one table serves all consumers instead of parallel copies.
