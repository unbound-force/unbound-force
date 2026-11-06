---
tag: review-dispatch
author: jay-flowers
category: pattern
created_at: 2026-11-05T14:34:59Z
identity: review-dispatch-20261105T143459-jay-flowers
tier: draft
---

Cross-cutting spec scenarios pin unevenly in command contract tests when template phrasing varies: RD-FR-056's old-plugin stranding scenario shares one fragment ("the loaded plugin predates durable proposals") across all five consumer commands, so it pins with a single loop-level assertNormalizedContains, but its failed-finalization scenario is phrased differently per command (uf.review-pr.md: "report them as stranded, never store improvised lessons"; uf.unleash.md: "Skip storage when the finalization failed"), so pinning it needs per-command clauses over each template's existing fragment — or a phrasing-alignment pass before a shared pin can exist. When mapping spec scenarios to contract-test pins, check phrasing uniformity across the consumer templates before assuming one loop-level assertion suffices. (PR 949 Item B, RD-FR-056 scenario 4.)
