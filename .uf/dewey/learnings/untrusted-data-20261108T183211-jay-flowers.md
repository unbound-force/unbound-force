---
tag: untrusted-data
author: jay-flowers
category: gotcha
created_at: 2026-11-08T18:32:11Z
identity: untrusted-data-20261108T183211-jay-flowers
tier: draft
---

Untrusted-data framing is a property of the content class, not the consumption route. When a new command flow moves ingestion of external text (PR diffs, review bodies, inline comments) from tool-delimited child sessions into the parent session, the data-only note must be restated at the new ingestion point — the dispatch tooling's untrusted-evidence delimiters do not carry over to a light-check path that reads the same text bare. Sibling flows in the same command (e.g. uf.review-pr Step 4c's bounded-untrusted-context note) are the internal precedent to copy, and byte-identical scaffold mirrors under internal/scaffold/assets/ must receive the same edit or scaffold drift tests fail.
