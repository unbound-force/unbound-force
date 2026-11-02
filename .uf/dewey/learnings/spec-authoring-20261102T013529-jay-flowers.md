---
tag: spec-authoring
author: jay-flowers
category: gotcha
created_at: 2026-11-02T01:35:29Z
identity: spec-authoring-20261102T013529-jay-flowers
tier: draft
---

When a spec enumerates an exhaustive disposition vocabulary for a rewrite of existing error-handling code, diff the enumeration against the current function's full return-path set before approving the contract. Spec-driven rewrites can silently drop I/O-failure handling: the disposition list in the spec came from the happy-path state matrix and renamed "already configured" to "unchanged" while omitting the existing "failed" disposition for read/write errors, and the task's test list covered only the four success paths — so an implementer following the spec verbatim would have regressed error handling with no test to catch it. Concrete case: ensureGitignore in internal/scaffold/scaffold.go returns action "failed" with detail on read errors (1436-1443) and write errors (1467-1473), but the spec's replacement contract for it listed only created/upgraded/unchanged/skipped. The check is mechanical: for every rewritten function, list its current return paths (including error returns) and require each to map to a disposition in the new contract, or to an explicit decision to drop it.
