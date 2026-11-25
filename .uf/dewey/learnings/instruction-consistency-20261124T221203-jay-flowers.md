---
tag: instruction-consistency
author: jay-flowers
category: gotcha
created_at: 2026-11-24T22:12:03Z
identity: instruction-consistency-20261124T221203-jay-flowers
tier: draft
as_of: "unbound-force@c9241969995f590083f3f56f0a0347168c4d56f9"
---

Adding a new exit source changes which exit rules apply. If an exit path emits tool-rendered text from a tool other than the command's canonical exit renderer (for example consolidate_dispatch's intent_deviation_exit in /uf.unleash), update the command's "NEVER improvise exit text — MUST call format_unleash_exit" rule in the same change. Otherwise the two instructions contradict each other, and agents may render the exit twice or rebuild the inputs by hand.
