---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T01:14:30Z
identity: issue-triage-20261102T011430-jay-flowers
tier: draft
---

Unbound Force review-dispatch context-exhaustion mitigations split into two complementary layers that can trip duplicate checks during triage: the delivery layer (issue #880 — RD-FR-032 manifest delivery plus on-demand read_diff_chunk retrieval for large diffs) and the assembly layer (issue #770 — per-run context budgets trimming prompts and sibling evidence). Both present the same symptom, oversized review-dispatch payloads, through different mechanisms. When triaging dispatch-scalability issues, identify which layer an issue targets before judging duplication; cross-link the pair as related rather than closing either as a duplicate.
