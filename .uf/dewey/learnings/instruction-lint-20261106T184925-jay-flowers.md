---
tag: instruction-lint
author: jay-flowers
category: pattern
created_at: 2026-11-06T18:49:25Z
identity: instruction-lint-20261106T184925-jay-flowers
tier: draft
---

Shingle-based near-duplicate detection needs a frequency cap at the candidate-generation stage, kept separate from the verification stage. In instruction-lint's checkDuplicates (.opencode/lib/instruction-lint.ts), candidate generation walks every shingle's full posting list — quadratic in boilerplate frequency — while verification independently recomputes full-shingle Jaccard per candidate pair. Skipping shingles whose posting list exceeds a cap only during candidate generation bounds the scan at O(n·k·cap) and preserves scoring semantics exactly; the only recall change is duplication families whose every shared shingle exceeds the cap, which cap sizing controls. High-frequency shingles carry no discriminative signal (the IR stop-word argument), so the cap is the standard retrieve-then-verify pattern: cheap capped retrieval, exact verification.
