---
tag: determinism
author: jay-flowers
category: gotcha
created_at: 2026-11-15T05:06:22Z
identity: determinism-20261115T050622-jay-flowers
tier: draft
---

Stale-vs-true-duplicate review classification (commit_id == head_sha string equality plus finding-set resolved/new/severity diff) is mechanical and must live in resolve_review_mode or a sibling tool, not in Step 8a prose. The incident on #1024 (SHA inequality present but agent still recommended skip at 0.44 confidence) shows prose classification cannot reliably fix #1055. Correct boundary: tool returns boolean true-duplicate/stale plus presentation inputs (prior verdict, short SHAs, resolved vs new lists); prose keeps dual-channel visibility, option reasoning, and recommendation wording. Any prose→tool migration must include DE-007 seam tests (registration surface, documented output shape, write→read round-trip) in the same increment, and should net reduce command tokens.
