---
tag: dewey
author: jay-flowers
category: gotcha
created_at: 2026-11-07T16:25:01Z
identity: dewey-20261107T162501-jay-flowers
tier: draft
---

Knowledge-capture truncation is self-propagating until the batch is committed and gated: PR #968's batch includes a learning documenting PR #949's truncated-learning defect (dewey-20261104T225833) that is itself truncated at the same 512-character preview cap — the knowledge store's only warning about the defect class was cut off at capture time, stayed uncommitted, and the identical defect recurred on PR #968 before any copy existed in-repo. Prevention is mechanical on two axes: capture lesson text from the dispatch artifact's payload.proposals (via read_dispatch_artifact), never consolidate_dispatch's 512-char-capped tool-result preview; and commit learning batches promptly so captured warnings exist in the repository before the defect can repeat. After completing truncated bodies in a fix commit, refresh the Dewey store copies so semantic retrieval serves the completed text rather than the truncated captures.
