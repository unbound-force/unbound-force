---
tag: review-process
author: jay-flowers
category: gotcha
created_at: 2026-11-05T16:04:35Z
identity: review-process-20261105T160435-jay-flowers
tier: draft
---

When verifying a PR-feedback claim against "current file content", first confirm the working tree is at the PR head. On unbound-force PR 949 feedback item I, the session working tree did not contain the PR's changes (no .opencode/test/read-dispatch-artifact.test.ts on disk; no readDispatchArtifact or read_dispatch_artifact registration in the working-tree .opencode/plugins/review-dispatch/index.ts at 9202 lines, while the PR head adds the reader around lines 6228-6414), so working-tree reads and greps silently verified the BASE state, not the change. The read_diff_chunk retrievals were complete (one chunk per file) and authoritative; verification of feedback on unmerged PRs should rest on the diff, and working-tree reads should be trusted only after confirming the checked-out commit matches the PR head.
