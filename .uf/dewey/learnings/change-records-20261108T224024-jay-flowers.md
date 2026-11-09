---
tag: change-records
author: jay-flowers
category: gotcha
created_at: 2026-11-08T22:40:24Z
identity: change-records-20261108T224024-jay-flowers
tier: draft
---

Retroactive change records go stale on DE-004 numbers when fix rounds land after the counts were taken: PR #959's record says divisor-determinism is 246 lines (shipped: 263 per the diff hunk) and divisor-linguist 236->203/202 (base checkout: 262; diff +26/-58 puts head at 230), and the two records disagree with each other (203 vs 202). Before updating a retrospective record, re-derive the numbers from the final head — run `instruction-lint --json --base <BASE> --changed` and paste the delta rows — because a record written from an intermediate commit bounds the wrong commit.
