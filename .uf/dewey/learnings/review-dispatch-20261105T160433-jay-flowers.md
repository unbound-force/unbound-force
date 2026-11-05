---
tag: review-dispatch
author: jay-flowers
category: pattern
created_at: 2026-11-05T16:04:33Z
identity: review-dispatch-20261105T160433-jay-flowers
tier: draft
---

A bounded-read TOCTOU guard (pre-read stat-size check plus in-loop cumulative-bytes check, as in readDispatchArtifact from unbound-force PR 949, RD-FR-057) cannot be covered by a static real-file fixture: a regular file that passes the stat check can never deliver more bytes than its stat size during the read, so only the stat branch is reachable for real. A "growing file" fixture is no substitute: with no seam between stat and read, a concurrent append races the stat, and when both branches render byte-identical fail-closed errors the test can neither choose nor assert which branch fired — it can pass via the stat branch with the in-loop guard deleted. Deterministic branch coverage requires a scripted fake FileHandle: a partial vi.mock of node:fs/promises open with importOriginal (the .opencode/test suite's established pattern — unleash-branch-gate, unleash-formatters, unleash-tasks, model-matrix-integration, uf-finale) whose stat reports a small regular file and whose read delivers cumulative bytes past the bound, or extracting the read loop into an exported handle-parameterized helper tested directly.
