---
tag: review-dispatch
author: jay-flowers
category: gotcha
created_at: 2026-11-24T22:12:01Z
identity: review-dispatch-20261124T221201-jay-flowers
tier: draft
as_of: "unbound-force@c9241969995f590083f3f56f0a0347168c4d56f9"
---

When a compact tool result carries a key the calling command routes on (for example consolidate_dispatch's intent_deviation_findings / intent_deviation_exit, which pick between the /uf.unleash exit and the fix loop), the byte-budget drop loop must never drop it silently. Either keep a routing scalar that is never dropped, or set a *_truncated flag. Otherwise an absent key reads as 'no such finding' and the agent takes the wrong branch. Also, when an intent-level tool absorbs a chain (dispatch_review_runs acquiring intent context internally), update every skill that documents the old chain, not just the commands. Explicit-input-wins semantics let stale skill prose override the tool's own acquisition.
