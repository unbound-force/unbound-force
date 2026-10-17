---
tag: review-dispatch
author: jay-flowers
category: gotcha
created_at: 2026-10-15T18:27:34Z
identity: review-dispatch-20261015T182734-jay-flowers
tier: draft
---

Dispatched review-council child agents run in the parent session's working directory, which may be checked out at the PR base commit rather than the reviewed head. In the PR #733 review (opsx/address-feedback-model-normalization, head 4fc68ea), the working tree matched base ae46169 — pre-migration command prose present at the old line numbers, the new CHANGELOG entry absent, and openspec/changes/address-feedback-model-normalization/ plus the 2026-10-15 Dewey learnings absent from disk — so working-tree greps verified the base state and nearly produced a false "inline algorithm still present / spec scenario violated" finding. Before treating grep or read results as head-state verification in a dispatched review, confirm the tree matches the reviewed head: check a file the diff ADDS (the new CHANGELOG entry or the new spec directory) — its absence proves the tree is not at head. Head-state content claims must rest on the caller-provided diff (authoritative, blob-hash-identified) or an explicit head checkout, never the ambient working tree. Base-state greps still carry value as negative evidence about the base: combined with the diff, the base-only presence of the normalization prose in uf.address-feedback.md proved the extraction arc complete at head.
