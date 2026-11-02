---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T04:10:14Z
identity: issue-triage-20261102T041014-jay-flowers
tier: draft
---

Triage checklist for retro-agent issues proposing new AGENTS.md guidance in unbound-force repos: (1) confirm the proposal is additive guidance, not a modification of a protected gate (coverage thresholds, severity definitions, CI flags, review limits, pinned dependency versions); (2) confirm the guidance is judgment-oriented — mechanical procedural prose belongs in an executable tool per the determinism gate, but a judgment heuristic (e.g., "scan for existing patterns and decide whether they cover the same assertion category") is compatible since detecting semantic equivalence of test logic is not feasible for a tool; (3) confirm existing convention packs do not already cover the guidance — go.md TC-006 recommends the table-driven style but does not mandate extending existing tables, which is why issue #609's pattern-reuse gap was real; (4) weigh the maintenance location: guidance added to a convention pack (go.md) requires syncing the embedded scaffold mirror under internal/scaffold/assets/ enforced by drift tests, while AGENTS.md is a single-file edit — the lower-burden path for repo-specific examples. Issue #609 (test-pattern reuse bullet for AGENTS.md Testing Conventions) passed all checks.
