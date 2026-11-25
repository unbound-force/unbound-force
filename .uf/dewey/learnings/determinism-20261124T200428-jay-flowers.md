---
tag: determinism
author: jay-flowers
category: pattern
created_at: 2026-11-24T20:04:28Z
identity: determinism-20261124T200428-jay-flowers
tier: draft
as_of: "unbound-force@c9241969995f590083f3f56f0a0347168c4d56f9"
---

Moving an acquisition into its own tool (for example acquire_intent_context) does not finish a DE-005 collapse if the command prose still tells the agent to forward the tool's output verbatim into another tool (dispatch_review_runs). That remaining forward is itself a judgment-free chain. When review state must reach every dispatch, have the dispatch aggregator acquire it internally, as dispatch_review_runs already does for sibling evidence. Also, a routing step that filters findings by category before choosing an exit variant needs the findings surfaced in consolidate_dispatch's compact result. Otherwise the agent has to page through read_dispatch_artifact and filter by hand.
