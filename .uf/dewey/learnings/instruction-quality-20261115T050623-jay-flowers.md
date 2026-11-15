---
tag: instruction-quality
author: jay-flowers
category: gotcha
created_at: 2026-11-15T05:06:23Z
identity: instruction-quality-20261115T050623-jay-flowers
tier: draft
---

A fix for a gate that misjudged a condition can reintroduce the same failure class through its own branch predicate. PR 1077 fixed /uf.review-pr's duplicate-review gate (issue #1055: a stale same-verdict review recommended for skip) by routing on "the finding set has not materially changed" — an undefined materiality threshold, so two runs on identical inputs can present opposite gate recommendations. The Linguist's test for gate prose: name every operand's source and give every branch condition an operational rule (e.g., "any finding added, removed, or severity-changed counts as material"), or encode the comparison in the owning tool where the rule is code.
