---
tag: test-architecture
author: jay-flowers
category: pattern
created_at: 2026-11-25T06:36:49Z
identity: test-architecture-20261125T063649-jay-flowers
tier: draft
as_of: "unbound-force@e8eba4889b2e87b6dd4870ef340fc27ed7985c8e"
---

For testing a tool whose contract is git/gh/filesystem choreography, inject fakes that share one mutable model — fetches add objects to the object store, worktree add/remove mutate the worktree list, removeDir is observable, now() and randomId() are deterministic — so ordering and no-side-effect claims become assertable without real git; then add exactly one real-git round trip to prove the seam end to end. Shipped in unbound-force review_head_root (.opencode/test/review-head-root.test.ts, 55 unit tests) plus preflight-declared-roundtrip.test.ts: the fakes pin "never calls checkout/switch/stash/reset", "no fetch or worktree before UNTRUSTED_HEAD", and "no FETCH_HEAD read", while the single round trip proves the user's HEAD, branch, git status, untracked file, and .git/config are byte-unchanged across prepare → declared checks → release.
