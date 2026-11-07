---
tag: dewey
author: jay-flowers
category: gotcha
created_at: 2026-11-07T16:35:32Z
identity: dewey-20261107T163532-jay-flowers
tier: draft
---

Cross-session Dewey learning batches bypass the consolidation contract's near-identical merge. duplicate_proposal_merged operates per-dispatch, so a retrospective PR aggregating lesson proposals from separate review sessions (unbound-force PR #968: 50 files, seven clusters persisting 22 near-identical files restating ~6 insights) reaches the graph with duplicates intact, where they remain retrieval noise until dewey_compile runs — and compile is not automatic (PR #815's draft learnings were still uncompiled on disk when PR #968 re-captured one). Batch capture needs an explicit cross-session dedup step before commit: a body-similarity scan across the batch being assembled plus a duplicate check against the existing graph — dewey_semantic_search when the embedding model is loaded, or a tag-directory scan of .uf/dewey/learnings/ when it is not (in this review semantic search was down and Dewey keyword search did not index the on-disk learning; the direct file read was the only working verification path). Triage guidance for dedup feedback on such batches: draft-tier near-identicals are the compile tier's designed input, so the ask is AUTHOR-DECIDES unless a MUST rule is violated; any collapse must complete truncated bodies first (diffstat hides mid-sentence cuts at +10/-0 per file) and fold members' unique details — distinct tags, precedent references, bound enumerations — into the kept statement rather than pure-deleting.
