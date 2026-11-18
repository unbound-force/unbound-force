---
tag: instruction-quality
author: jay-flowers
category: pattern
created_at: 2026-11-16T11:39:27Z
identity: instruction-quality-20261116T113927-jay-flowers
tier: draft
as_of: "unbound-force@474f5aee31d7e0bc4583635cccfeec6030277b33"
---

When reviewer feedback asks to extract command prose into a tool (DE-005 class), scope the acceptance as a coordinated artifact sweep, not a two-file edit: the prose-mandating contract usually lives in more artifacts than the feedback names. On PR #1157 (per-task-delegation-budgets), the never-blank-retry mandate was hard-bound to command prose in three artifacts added by the same change — the spec FR body (PTB-FR-003), the design decision (D3), and the command prose itself — while the feedback named only the spec FR; a literal application would have left design.md mandating a mechanism the implementation no longer contained. Verification triad before accepting such feedback: (1) enumerate the flagged prose's branches to confirm it is judgment-free; (2) confirm the target tool already holds every input the chain needs (original prompt, structured result, capture seams, retryable classification); (3) grep the change's spec and design artifacts for the mandate's wording — every hit must be amended in the same change, and the amended requirement must keep acceptance observable at the tool's seam.
