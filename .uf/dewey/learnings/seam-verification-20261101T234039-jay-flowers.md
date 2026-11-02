---
tag: seam-verification
author: jay-flowers
category: pattern
created_at: 2026-11-01T23:40:39Z
identity: seam-verification-20261101T234039-jay-flowers
tier: draft
---

A census-count disagreement between two records is resolved by enumerating the sites each count was computed against, never by choosing between the printed numbers — disagreeing counts usually reflect different site sets, not typos. Verified while triaging unbound-force issue #823 (2026-11-01): golden-output-shape.test.ts:1036 says "review-dispatch registry -- 14 documented sites" while the describe at :1112 says "(15 documented sites)"; enumerating the registry's site labels yields exactly 15 — the header-listed 14 plus capture_review_diff [uf.unleash.md:620], a site absent from the per-plugin header comment, from the "44 of the 48" coverage arithmetic, and from the metrics.md census entirely. The 47-vs-48 drift has the same shape: metrics.md's 47 predates uf_finale_base_state (uf.finale.md:132) and fetch_pr_review_context (uf.address-feedback.md:172), both documented after the task-2.1 census without a census event. When sites ship without census events, the recounted total exceeds both printed numbers — rebuild counts from grep-enumerated ground truth (sites and label anchors across golden suite, sibling suites, and the registry) and land all corrections in one commit.
