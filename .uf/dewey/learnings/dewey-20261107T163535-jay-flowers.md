---
tag: dewey
author: jay-flowers
category: gotcha
created_at: 2026-11-07T16:35:35Z
identity: dewey-20261107T163535-jay-flowers
tier: draft
---

Retrospective Dewey learning batches captured across sessions bypass the consolidation contract's near-identical merge: duplicate_proposal_merged (RD-FR-055 design D2) compares proposals within a single dispatch only — never across capture sessions or against the existing graph, and tags never participate — so a multi-session batch can both re-capture an on-main learning (PR #968: scaffold-mirrors-20261105T174838 restated learning/scaffold-mirrors-20261025T145313 from PR #815) and persist seconds-apart restatements of one insight as separate files. Pre-push dedup for multi-session batches: (1) for each file, search the existing graph by tag plus a distinctive body phrase — a match on both the referenced-PR pair and the core insight marks a re-capture; drop it or rewrite it as an explicit delta referencing the prior identity. (2) Within the batch, group by tag and compare bodies: same-tag clusters differing only in phrasing collapse to the fullest statement; cross-tag variants carry a trade-off because a dewey learning exposes a single tag, so collapsing them narrows retrieval surfaces — retain them or record the choice. Severity anchor: a pure re-capture is MEDIUM zero-waste (duplicate graph node); intra-batch near-identicals are LOW-to-MEDIUM retrieval noise the draft/compile tier absorbs by design.
