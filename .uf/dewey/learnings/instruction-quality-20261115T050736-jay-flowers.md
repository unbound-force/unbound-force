---
tag: instruction-quality
author: jay-flowers
category: gotcha
created_at: 2026-11-15T05:07:36Z
identity: instruction-quality-20261115T050736-jay-flowers
tier: draft
---

In uf.review-council.md the word family 'truncat-' denotes two distinct mechanisms: host-side truncation of tool results (Section 4 truncation recovery / spill files) and child output-budget truncation (finish reason 'length', in the consolidation-advisory prose). Instruction prose that names a truncation effect in that file should name the truncation source explicitly; a bare 'truncated' invites cross-mechanism misreads by executing agents.
