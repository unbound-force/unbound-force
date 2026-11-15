---
tag: compaction-storm-on-a-tight-context-mode
author: jay-flowers
category: pattern
created_at: 2026-11-15T11:40:33Z
identity: compaction-storm-on-a-tight-context-mode-20261115T114033-jay-flowers
tier: draft
---

Delivery retrospective for opsx/compaction-storm-on-a-tight-context-mode (issue #1079, completed 2026-11-15): three bounded behavioral fixes in the delegation layer (headroom gate in model selection, compaction-storm detection failing fast to retryable context_thrash with a resume envelope, final-text success guard for compaction summaries and synthetic continue prompts) plus two fail-closed knobs (selection.footprint_margin_tokens 1000-256000 default 24000; delegation.storm_threshold 2-10 default 3). Process lessons: (1) review-fix iterations that edit instruction files after the change record was authored must update the record with the instruction-lint delta rows (DE-004) — the delta evidence (.uf/scripts/instruction-lint --json --base <merge-base> --changed) is cheap to capture and belongs in the CHANGELOG entry or PR body; (2) single-sourcing compaction accounting in delegation-budget.ts (foldProbeObservations) kept the storm detector and the idle watchdog from drifting apart; (3) the write→read round-trip test through get_progress is the cheapest proof that a resume envelope actually supports orchestrator re-delegation.
