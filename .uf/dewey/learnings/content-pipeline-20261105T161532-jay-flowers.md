---
tag: content-pipeline
author: jay-flowers
category: pattern
created_at: 2026-11-05T16:15:32Z
identity: content-pipeline-20261105T161532-jay-flowers
tier: draft
---

Documentation-gate duplicate checks must verify a candidate issue's change scope by reading its body, not just its title keywords. In the unbound-force repos, docs tracking issues are filed per-change as in-repo + website pairs (#740+website#310, #791+website#317, #810+website#320, #856+website#340), so title-keyword matches for a new change almost always belong to a prior change. During the PR #948 feedback review (2026-11-05), keyword searches across unbound-force/unbound-force and unbound-force/website surfaced seven candidate docs issues; reading each body's change/PR reference showed every one scoped to an earlier change, establishing that no issue tracked the current change. Establishing "no docs issue exists for THIS change" therefore requires reading candidate bodies, and the finding should cite the nearest prior-change issue pair as the precedent the new change must follow. A change's linked bug-report issues are not docs tracking issues.
