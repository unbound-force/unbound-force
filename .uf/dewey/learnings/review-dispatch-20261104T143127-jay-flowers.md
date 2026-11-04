---
tag: review-dispatch
author: jay-flowers
category: pattern
created_at: 2026-11-04T14:31:27Z
identity: review-dispatch-20261104T143127-jay-flowers
tier: draft
---

Byte-size ceilings in tools are cheaply testable without writing large files: fs.truncate (node:fs/promises) extends an existing small fixture to limit+1 bytes and POSIX ftruncate leaves a hole (sparse on APFS/ext4), so an lstat-based size check rejects before any read — e.g. PR 949's read_dispatch_artifact ARTIFACT_MAX_FILE_BYTES (64 MiB) branch at .opencode/plugins/review-dispatch/index.ts:6285, asserted as {status: "error", code: "invalid_artifact"}. Generalizable rule: when a spec MUST clause enumerates fail-closed conditions, map every enumerated condition to a named test, because (a) the scenario block may list fewer conditions than the requirement body — RD-FR-057's "Reader fails closed" scenario omits the 64 MiB ceiling its own MUST clause enumerates, and the suite verified the scenario list instead of the clause — and (b) aggregate coverage floors (94% statements, 85% branches in .opencode/vitest.config.ts) never surface a single uncovered branch. Reinforces learning review-dispatch-coverage-20261009T031707-jay-flowers.
