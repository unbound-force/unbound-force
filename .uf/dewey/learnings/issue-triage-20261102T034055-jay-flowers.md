---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T03:40:55Z
identity: issue-triage-20261102T034055-jay-flowers
tier: draft
---

When triaging instruction-consolidation issues (moving protocol prose repeated across multiple command files into one shared skill), verify two things beyond the issue's own claims. First, diff the duplicated copies against each other: in unbound-force/unbound-force issue #663 the triplicated dispatch protocol had already diverged — uf.address-feedback.md Guardrail 11 mandates execution through dispatch_agent_run while the same command's Section 2.3.4 and the dispatch-advisor skill's Invocation Tool section mandate dispatch_review_runs. Materialized divergence both proves the issue's divergence-risk premise and identifies which copy is authoritative for the consolidated text. Second, verify each proposed lifecycle step against the current tool surface: the issue's steps named dispatch_agent_run, prepare_lesson_learning, and finalize_review_dispatch, but the three target commands now execute through dispatch_review_runs (the DE-005 intent-level aggregate) and dual-mode consolidate_dispatch. A skill expansion must bind lifecycle stages to the tools the commands actually invoke, using the issue's no-behavioral-change acceptance criterion as the tiebreaker when the proposal's tool names have gone stale.
