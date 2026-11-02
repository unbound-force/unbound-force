---
tag: docs-synchronization
author: jay-flowers
category: gotcha
created_at: 2026-11-01T23:21:03Z
identity: docs-synchronization-20261101T232103-jay-flowers
tier: draft
---

In unbound-force/unbound-force, the /uf.unleash autonomous-pipeline description is duplicated across docs/usage.md (Go Fully Autonomous section, exit clause at lines 179-181) and docs/architecture.md (Autonomous Mode section, exit clause at lines 453-455 as of 2026-11-01), with each copy line-wrapped independently — the architecture.md copy wraps mid-phrase, so grepping the full quoted sentence finds only the usage.md copy. When verifying or implementing docs issues that quote this passage: search for the shortest stable fragment (e.g. "encounters ambiguity, review"), expect cited line numbers to drift between filing and triage (issue #840's citations drifted ~27/~13 lines within five days), and treat any /uf.unleash clarify/exit-semantics change as requiring lockstep edits to both copies plus consolidation with overlapping open docs issues (#739, #840).
