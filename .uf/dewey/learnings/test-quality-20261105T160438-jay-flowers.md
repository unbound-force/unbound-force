---
tag: test-quality
author: jay-flowers
category: pattern
created_at: 2026-11-05T16:04:38Z
identity: test-quality-20261105T160438-jay-flowers
tier: draft
---

An in-loop size guard paired with a pre-read stat check is unreachable by static fixtures: an unchanged file never delivers more bytes than its stat-time size, so only the pre-check branch ever executes and the in-loop TOCTOU guard silently loses all regression protection (verified on PR 949: readDispatchArtifact's 64 MiB live-stream bound — the sparse-truncate test covered the stat branch only). Deterministic vitest recipe: partial-mock node:fs/promises via vi.mock with importOriginal, then override open for exactly one call (mockImplementationOnce) to return a synthetic FileHandle whose stat() passes the pre-check and whose read() fills the buffer until it has delivered slightly more than the cap, then returns bytesRead 0 — the real loop trips the in-loop bound deterministically, and if the guard is ever removed the read completes and fails a message assertion (parse error instead of size error) instead of hanging. Pin the branch by the fixture's small stat size (the pre-check cannot fire) or a distinguishable error message, since pre-check and in-loop guards commonly return identical code and message. The real-fs alternative — sparse-truncate just under the cap, then truncate past it from a timer while the loop reads — is feasible (the loop needs ~1023 sequential 64 KiB reads, a real timing window) but is timing-dependent and can flake under CI load.
