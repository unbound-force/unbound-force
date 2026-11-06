---
tag: openspec
author: jay-flowers
category: pattern
created_at: 2026-11-06T14:48:10Z
identity: openspec-20261106T144810-jay-flowers
tier: draft
---

Before a change moves a value that specifications hardcode (a pinned SHA, version, threshold, flag), grep both the old and the new value across openspec/changes/*/specs/ for unarchived sibling deltas that assert the value inside a MUST requirement — including deltas of other capabilities, not just the capability being modified. A sibling delta can hardcode an already-stale value for the same reference (gaze-ai-report's specs/ci-local pinned the Gaze reusable workflow at c133b118 while ci_local.yml had already moved to 8bdfb59), so grepping only the value being replaced misses it. Update the sibling delta's hardcoded value in the same PR when it is directly tied to the value being moved, or record the divergence in the change's design with a pre-archive follow-up; otherwise the deltas land in canonical specs as contradictory MUSTs, and the stale value can later drive a spec-directed revert that re-breaks an enforced gate.
