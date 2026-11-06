---
tag: instruction-lint
author: jay-flowers
category: pattern
created_at: 2026-11-06T18:49:22Z
identity: instruction-lint-20261106T184922-jay-flowers
tier: draft
---

Shingle-based near-duplicate detection must bound high-document-frequency shingles or runtime grows quadratically with corpus size: a shingle present in m documents is walked m times and admits ~m^2 candidate pairs while carrying no discriminative signal at high Jaccard thresholds. Fix: skip shingles whose posting list exceeds a cap during candidate generation (keep capped shingles in the similarity scoring) — a pair meeting a 0.85 Jaccard threshold necessarily shares low-frequency shingles, so detection quality is preserved. Verified on unbound-force PR #958 (.opencode/lib/instruction-lint.ts checkDuplicates; advisory pre-flight check with a 120-second window).
