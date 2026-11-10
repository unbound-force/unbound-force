---
tag: seam-verification
author: jay-flowers
category: gotcha
created_at: 2026-11-08T18:14:57Z
identity: seam-verification-20261108T181457-jay-flowers
tier: draft
---

Argv-assertion mock tests pin invocation shape, not git executor semantics — and git subcommands diverge on unmatched pathspecs: `git add`, `git rm`, `git checkout --`, and `git restore` die with "pathspec did not match", while `git reset -q -- <path>` is believed (not verified) to no-op silently. When a plugin tool adds a load-bearing git invocation whose common path is an unmatched pathspec — e.g. uf_finale_stage's learnings reset, where the spec's primary scenario (untracked learnings) guarantees no index entries match — the behavior contract exists only as assumption until a Layer 3 real-repo round-trip test pins it: mkdtemp scratch repo, execFileSync git through the real factory (exec, cwd) contract, assert exit status and index state for the no-match, untracked, and pre-staged cases. Failure mode without the test: execFileSync throws on non-zero exit and the tool collapses to a non-retryable error envelope on every run for the affected consumer class. Precedent: .opencode/test/review-diff.test.ts "real-repo write->read round-trip" (review-diff.test.ts:361).
