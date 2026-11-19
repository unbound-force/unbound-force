---
tag: determinism
author: jay-flowers
category: pattern
created_at: 2026-11-14T23:18:30Z
identity: determinism-20261114T231830-jay-flowers
tier: draft
---

DE-001 boundary for hybrid gate steps (from the quick/fix-review-pr-stale-duplicate-check review, Nov 2026): when a command gate's branch predicate mixes a deterministic fact (e.g., prior review commit_id == frozen head_sha) with a judgment evaluation (e.g., "finding set has not materially changed"), the step classifies as judgment and stays in prose. DE-001's MUST extraction targets steps that are fixed procedures end-to-end (twice-stable output) and prose that duplicates a classification a tool already renders — not fact conjuncts inside a human-confirmation gate. Shipped precedent for the deterministic-only variant: resolve_review_mode owns the all-deterministic light/full re-review classification, and uf.review-pr Step 3.11 forbids restating or recomputing it. Corroborating signal: instruction-lint's IL-041 heuristics flag multi-command fenced bash blocks, not inline comparisons inside gates (no IL-041 fired on the changed lines). Opportunistic migration remains open: the Step 3.10-i review fetch (jq projection) or a resolve_review_mode sibling could render a per-review commit-match flag for gate prose to consume, keeping judgment at the tool boundary per DE-006.
