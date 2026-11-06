---
tag: dewey-learnings
author: jay-flowers
category: gotcha
created_at: 2026-11-06T20:02:58Z
identity: dewey-learnings-20261106T200258-jay-flowers
tier: draft
---

Bulk-committed dewey learning files can be silently truncated mid-payload: frontmatter and file count stay intact while the single body line ends mid-word, losing the learning's concrete conclusion. Observed in unbound-force PR #957 — all nine learnings from the 2026-11-04T22:49 batch (supply-chain-20261104T224929 through regression-protection-20261104T224939) end mid-sentence ("the rewritte", "unbound-forc", "tasks.md for "), while the 2026-11-05/06 batches are complete. Detection is a cheap pre-commit grep over .uf/dewey/learnings for body lines lacking terminal punctuation — body lines start uppercase while frontmatter lines start lowercase, so three patterns cover it: ^[A-Z].*[a-z]$ (mid-word), \S $ (trailing space), and ^[A-Z].*[0-9]$ (digit ending); sentence-complete bodies match none. Treat any match as a truncated store and re-store the learning with its full text before committing.
