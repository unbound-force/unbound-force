---
tag: review-dispatch
author: jay-flowers
category: gotcha
created_at: 2026-11-04T15:01:38Z
identity: review-dispatch-20261104T150138-jay-flowers
tier: draft
---

Applying a review-feedback fix to .opencode/plugins/review-dispatch/index.ts requires patching its byte-identical scaffold mirror internal/scaffold/assets/opencode/plugins/review-dispatch/index.ts in the same increment — TestEmbeddedAssets_MatchSource (internal/scaffold/scaffold_test.go) fails make test on divergence, and uf init deploys the mirror to every hero repo, so a one-sided fix either breaks the build or ships the unhardened pattern org-wide. When the fix swaps guard mechanics (PR 949 feedback item 7: lstat-then-readFile to open with O_NOFOLLOW + fstat for readDispatchArtifact's RD-FR-057 fail-closed contract), keep the documented error codes so existing tests stay green (ELOOP from the O_NOFOLLOW open maps through the non-ENOENT catch to invalid_artifact, which the symlink test asserts; ENOENT stays not_found), keep the guard-before-read ordering that the sparse-truncation oversized-artifact test relies on, handle fs.constants.O_NOFOLLOW being undefined on some platforms, add O_NONBLOCK or an equivalent so the open cannot block on a FIFO where lstat's isFile() previously rejected it, and test every new branch — the .opencode suite's 94 percent statements floor is a report-and-stop gate per the change's own design Risks.
