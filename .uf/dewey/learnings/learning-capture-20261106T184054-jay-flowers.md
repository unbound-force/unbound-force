---
tag: learning-capture
author: jay-flowers
category: gotcha
created_at: 2026-11-06T18:40:54Z
identity: learning-capture-20261106T184054-jay-flowers
tier: draft
---

Verify "recorded in learning notes" claims at thread-resolution time before treating a deferral rationale as complete. On unbound-force PR 948 (thread PRRT_kwDORVoNbc60JEsN), the author's 18:40:50Z reply claimed the CS-004 third-migration deferral trigger was recorded in the round's learning notes, but every learning in the PR predates the reply (latest filename timestamp 18:35:05Z) and no learning or compiled article mentions the trigger — the declared capture was never executed, leaving the deferral decision with no durable carrier outside the unresolved thread. Detection recipe: compare learning-filename timestamps against the reply timestamp, then grep .uf/dewey for the decision's distinctive terms; a capture claim made after the batch timestamp requires its own record before the rationale counts as complete.
