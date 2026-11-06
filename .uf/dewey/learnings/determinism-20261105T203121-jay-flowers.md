---
tag: determinism
author: jay-flowers
category: pattern
created_at: 2026-11-05T20:31:21Z
identity: determinism-20261105T203121-jay-flowers
tier: draft
---

Triage heuristic for reviewer feedback citing the determinism gate against command prose (precedent: PR #955 feedback item 7): validate the rule claim against determinism.md DE-001/DE-005 — git choreography is explicitly named mechanical instruction, and a judgment-free tool-call sequence must collapse into one intent-level tool — and confirm feasibility via the plugin-layer precedents (uf_unleash_branch_gate, uf_unleash_detect_resume, capture_review_diff, resolve_base_ref). Then compare the remedy's file footprint against the feedback item's immutable file inventory. When the extraction requires artifacts beyond the inventory — a new plugin tool, unit tests, command+mirror collapse, contract-test clause updates, and DE-007 same-increment seam verification — disposition the item AUTHOR-DECIDES with a tracked follow-up change rather than ACCEPT-in-PR scope expansion, especially when the reviewer's own review body labels the item longer-term and another dispatched item already recorded the follow-up. Adversary bonus: extraction to an argument-array tool also removes the shell-interpolation surface created by interpolating PR-derived refnames (head_ref) into command-prose shell templates.
