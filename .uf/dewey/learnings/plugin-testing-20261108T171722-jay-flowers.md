---
tag: plugin-testing
author: jay-flowers
category: gotcha
created_at: 2026-11-08T17:17:22Z
identity: plugin-testing-20261108T171722-jay-flowers
tier: draft
---

Mock-verified git command issuance does not verify git pathspec semantics. When a plugin tool introduces a new git command shape (e.g. `git reset -q -- <path>` alongside a pathspec-excluded `git add`), the mocked GitExec unit tests prove only that the right argv was issued — not that git accepts it in the common case. `git reset -- <pathspec>` can die with 'pathspec did not match any files' when nothing under the path is in the index, which is the default state on a fresh checkout. Plugin tools that add new git choreography need at least one real-git integration test (temp repo) covering the no-match/common path, the path-present-untracked path, and the pre-staged path; the registration-only smoke layer (scratch-smoke) and golden-shape mocks do not observe git behavior at all.
