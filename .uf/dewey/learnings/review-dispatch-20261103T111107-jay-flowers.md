---
tag: review-dispatch
author: jay-flowers
category: gotcha
created_at: 2026-11-03T11:11:07Z
identity: review-dispatch-20261103T111107-jay-flowers
tier: draft
---

OpenCode plugin tool contracts carry a documentation surface that plugin tests do not guard: the tool() description and the args describe() block. unbound-force PR #929 added a detail parameter to consolidate_dispatch's zod schema plus exported builder functions (buildSummaryRetrieval, buildFullRetrieval, buildCompactFinalization) that the new summary-retrieval.test.ts imports and tests directly — but the registered tool's description still documented the pre-change contract ("returns the payload JSON"), leaving the breaking compact-finalization shape and the summary mode undiscoverable by calling agents outside the five updated commands.
