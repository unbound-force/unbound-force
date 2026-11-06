---
tag: instruction-lint
author: jay-flowers
category: pattern
created_at: 2026-11-06T20:25:04Z
identity: instruction-lint-20261106T202500-jay-flowers
tier: draft
---

Shingle-based near-duplicate detection must bound high-document-frequency shingles at the candidate-generation stage, kept separate from the verification stage: a shingle present in m paragraphs is walked m times and admits ~m^2 candidate pairs, while verification independently recomputes full-shingle Jaccard per pair. Skipping shingles whose posting list exceeds a cap (e.g. 64) only during candidate generation bounds the scan at O(n·k·cap) and preserves scoring semantics exactly — high-frequency shingles carry no discriminative signal at a high Jaccard threshold with a word-count floor (0.85 over 40+ words), so the cap costs no true positives (the IR stop-word argument; the standard retrieve-then-verify pattern). Two edge behaviors must be designed for and pinned by tests: (1) verbatim text replicated into more files than the cap has every shared shingle above the cap and escapes clustering entirely — pair the capped scan with an O(n) exact-duplicate pass keyed on the sorted shingle-set (or normalized-text) hash, so byte-identical copies stay clusterable; and (2) the recall cliff: near- (not byte-) identical copies beyond the cap share only over-cap shingles, generate no candidates, and report nothing — accepted boilerplate filtering, but only acceptable when a test pins the behavior and a comment names it (instruction-lint .opencode/lib/instruction-lint.ts checkDuplicates, MAX_POSTING = 64, advisory pre-flight check with a 120-second window on PR #958).
