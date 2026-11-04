---
tag: review-dispatch
author: jay-flowers
category: pattern
created_at: 2026-11-04T14:31:22Z
identity: review-dispatch-20261104T143122-jay-flowers
tier: draft
---

read_dispatch_artifact enforces its 64 MiB ARTIFACT_MAX_FILE_BYTES bound by lstat before readFile, so the rejection branch is testable without materializing 64 MiB: fs.truncate an existing valid artifact to 64*1024*1024 + 1 bytes (sparse on APFS/ext4/tmpfs) and assert {status: "error", code: "invalid_artifact"} — the size check fires before any read. Generalizes to any lstat/stat-based size guard that precedes the read: test it with sparse truncation at limit+1 (the boundary strictly greater than the limit, since the check is info.size > limit). Remember that the .opencode suite's 94% statements and 85% branches floors never surface a single uncovered error branch — size-limit and I/O-catch branches in the review-dispatch plugin have repeatedly needed dedicated tests (prior instance: learning/review-dispatch-coverage-20261009T031707-jay-flowers, MAX_PROMPT_BYTES branch at 84.93% vs an 85% branch floor).
