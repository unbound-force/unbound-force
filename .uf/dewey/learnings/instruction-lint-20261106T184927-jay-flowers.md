---
tag: instruction-lint
author: jay-flowers
category: pattern
created_at: 2026-11-06T18:49:27Z
identity: instruction-lint-20261106T184927-jay-flowers
tier: draft
---

Near-duplicate detection that generates candidate pairs by walking shingle posting lists degrades quadratically once boilerplate phrasing accumulates: a shingle present in m paragraphs costs m^2 scan work because each containing paragraph walks the full m-entry list. Bound the walk — skip shingles whose posting list exceeds a fixed cap (e.g., 64). At a high Jaccard threshold with a word-count floor (0.85 over 40+ words, i.e., 36+ shingles per paragraph), high-frequency shingles cannot lift a pair over threshold, so the cap costs no true positives. Edge case to design for: verbatim text replicated into more files than the cap has every one of its shingles above the cap and escapes clustering entirely; pair the capped scan with an O(n) exact-duplicate grouping keyed on normalized-text hash, or set the cap generously, so widely replicated copies are still flagged.
