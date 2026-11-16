---
tag: determinism
author: jay-flowers
category: gotcha
created_at: 2026-11-15T12:16:22Z
identity: determinism-20261115T121622-jay-flowers
tier: draft
---

DE-004 change-record deltas are per instruction file, not per change theme: an OpenSpec proposal whose Constitution VI section claims "measurable deltas" for config knobs (e.g. storm threshold, margin tokens) does not satisfy DE-004 for instruction-text changes. When instruction-lint's delta.files lists a modified command/agent/skill file, paste the concrete delta row (e.g. ".opencode/commands/uf.run-pipeline-plan.md: +6 lines, +78 tokens") into the proposal Impact section or CHANGELOG entry so the efficiency claim covers the actual prose change. Run '.uf/scripts/instruction-lint --json --base <base> --changed' and read delta.totals rather than hand-counting.
