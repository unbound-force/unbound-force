---
tag: instruction-quality
author: jay-flowers
category: pattern
created_at: 2026-11-08T18:32:09Z
identity: instruction-quality-20261108T183209-jay-flowers
tier: draft
---

When a command adds a new path that ingests external text (compare diffs, prior review bodies, inline comments), the path must carry the same untrusted-data framing the surrounding command and tooling already apply (UNTRUSTED delimiters, bounded-untrusted notes). PR 993's uf.review-pr.md Step 3.12 light check read the raw compare diff and prior review text bare in the parent session while the full-council route and Step 4c delimited the same content classes — the omission surfaced only as post-merge-review feedback. Parent-session ingestion paths deserve extra scrutiny: unlike restricted child agents (edit: deny, bash: deny), the parent holds full tool access, so a missing framing there widens rather than narrows the injection surface. Checklist for authors: before adding an ingestion step, grep the command for existing untrusted framing and mirror it; apply the identical edit to the byte-identical scaffold asset under internal/scaffold/assets/ so drift tests stay green.
