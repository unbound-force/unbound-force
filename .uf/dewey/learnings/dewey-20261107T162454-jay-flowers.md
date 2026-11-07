---
tag: dewey
author: jay-flowers
category: gotcha
created_at: 2026-11-07T16:24:54Z
identity: dewey-20261107T162454-jay-flowers
tier: draft
---

Learning-batch body completeness is mechanically checkable at review time even though diffstat hides it: every .uf/dewey/learnings/*-jay-flowers.md file is 8 frontmatter lines plus one body line (uniform +10/-0 in diffstat), so diffstat shape says nothing about integrity — read the end of each body. A body ending without terminal punctuation, or sitting at the ~512-character content cap, marks a capture from consolidate_dispatch's pre-consumer-bounds proposals preview, which cut each proposal's information to 512 characters and persisted no full copy (issue #941: the cut copy was the only copy). PR #968 shipped 15 such bodies of 50 (7 from the 2026-11-03 21:18 capture session, 8 from 2026-11-04 22:58) — the second consecutive batch after PR #949's — and one truncated file (dewey-20261104T225833) was itself the learning documenting the defect, confirming the capture path truncated content-agnostically. Pre-#949 captures are structurally unrecoverable from dispatch artifacts (the artifact carried no proposals; the authoring sessions may also live on another machine), so the accepted remediation is cd603598's pattern: preserve the verified prefix verbatim, complete the sentence from the public source finding with terminal punctuation, and record which source comment grounded each completion. Post-#949 captures persist full payload.proposals; the remaining guard is a mechanical completeness check gating learning-batch commits.
