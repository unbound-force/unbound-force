---
tag: instruction-quality
author: jay-flowers
category: gotcha
created_at: 2026-11-16T08:19:29Z
identity: instruction-quality-20261116T081929-jay-flowers
tier: draft
---

Pinned prose mirrored across files that run in different session contexts cannot carry context-specific channel selection. A disjunctive channel list like "obtain a human decision via the `question` tool or the orchestrator relay" reads as both-channels-work in every copy, but question prompts do not surface from parented child sessions — a spawned subagent or delegated child that picks the `question` option wedges at the gate until delegation timeout, after which gate-timeout recovery converges on the same decision. When one copy executes exclusively in a child context, put the context selection in adjacent prose OUTSIDE the pinned bytes (the pin test only requires the block verbatim; adjacent sentences are compatible) instead of diverging the pinned text.
