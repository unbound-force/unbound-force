---
tag: dewey-learnings
author: jay-flowers
category: gotcha
created_at: 2026-11-06T22:14:58Z
identity: dewey-learnings-20261106T221458-jay-flowers
tier: draft
---

Existence-verification does not cover count-claims. The 2026-11-06 coffee-import reconciliation (PR #979) corrected the false ".uf/dewey/compiled/ does not exist" assertion by verifying the directory exists, but left the corpus size restated three different ways across the four files citing it — "40 articles" (three files, counting the directory's 40 entries including _index.md) and "20-plus-article" (one file) against a ground truth of 39 articles plus the index. When a reconciliation pass verifies a directory-backed claim, enumerate the directory, distinguish index/manifest files from content files, record one canonical count, and grep the change's own added files for every restatement of that count before commit — FA-002 numerical consistency applies within a single commit's files, not only across published content.
