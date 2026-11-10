---
tag: determinism
author: jay-flowers
category: gotcha
created_at: 2026-11-08T18:04:14Z
identity: determinism-20261108T180414-jay-flowers
tier: draft
---

DE-001 admits no design-note deferral for new instruction. When a spec's design.md labels a tool extraction "a natural follow-up", the prose still violates DE-001 at merge unless the same increment ships the tool — the determinism pack's "existing prose is migrated opportunistically at review time" clause covers only pre-existing prose, never instruction added in the same change, and a tasks.md entry or tracked issue binds a follow-up without exempting the shipped prose. Verified on PR 993 (uf.review-pr Step 3.11): design.md recorded the deferral, tasks.md had no binding task, and the council filed the CRITICAL anyway.
