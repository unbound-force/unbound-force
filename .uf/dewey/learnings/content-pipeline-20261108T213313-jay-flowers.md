---
tag: content-pipeline
author: jay-flowers
category: gotcha
created_at: 2026-11-08T21:33:13Z
identity: content-pipeline-20261108T213313-jay-flowers
tier: draft
---

Docs/blog/tutorial issues filed in a later review round of the same change must re-run the open-issue search before filing: round-2 filings duplicate round-1 filings when the search-first step is skipped. PR 959 filed in-repo #980 and website#383 in round 1 (Nov 6), then round 2 (Nov 8) filed in-repo #995 and website#394 for the same scope — no labels on the round-2 filings, no cross-references between any of the four — fragmenting the docs gate. Remediation is consolidation plus labels (gh issue edit <N> --add-label docs) and mutual references, and the proposal should record the final canonical issue set so verification resolves against one set.
