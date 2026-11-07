---
tag: dewey
author: jay-flowers
category: pattern
created_at: 2026-11-07T16:35:26Z
identity: dewey-20261107T163526-jay-flowers
tier: draft
---

Dedup feedback on Dewey learning batches is verifiable and repairable with graph evidence, and the fix differs by cluster kind. Before advising a collapse, keyword-search the Dewey graph for each cluster's distinctive tokens: on unbound-force PR #968 item 17, the reviewer reported one re-captured learning (PR #815), but graph search surfaced three prior on-graph statements of the same blob-hash mirror-verification insight (learning/scaffold-mirrors-20261030T212223, learning/review-technique-20261029T155628, learning/documentation-pipeline-20261020T000306), so the PR's cluster-1 files were pure re-captures deletable outright — the graph already held complete statements. Two fix shapes: cross-session re-captures of already-stored insights delete; same-batch near-identicals collapse to the fullest complete statement, with the survivor chosen by completeness, not recency — two of three cluster-1 files ended truncated mid-sentence, and three item-17 files overlapped the sibling truncation-feedback item, so bodies must be completed before or during collapse or the survivor permanently loses the security pattern's remediation tail. Cluster membership itself is verifiable mechanically from the diff: created_at timestamps seconds apart plus shared distinctive tokens; the reviewer's 7-cluster/22-file count decomposed as the parent's 6 enumerated clusters (18 files) plus a mocked-FileHandle stream-guard cluster (review-dispatch-20261105T160430/33 + test-quality-20261105T160438/174850). Structural cause: consolidate_dispatch's near-identical merge (duplicate_proposal_merged) runs per-dispatch across personas, never across capture sessions.
