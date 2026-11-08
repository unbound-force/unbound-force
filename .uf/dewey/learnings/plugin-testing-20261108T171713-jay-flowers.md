---
tag: plugin-testing
author: jay-flowers
category: pattern
created_at: 2026-11-08T17:17:13Z
identity: plugin-testing-20261108T171713-jay-flowers
tier: draft
---

Argv-asserting unit tests (mocked exec) plus registration-only smoke tests leave new git command shapes unpinned against real git. When a plugin tool gains a new git invocation whose pathspec may match nothing in the consumer steady state, add a DE-007 Layer 3 real-git round-trip test: create a scratch repo, set up the state the policy makes universal (e.g. untracked-only .uf/dewey/learnings/ files after learnings moved off PR branches), invoke the tool through the real executor, and assert both success and the intended index state. Observed on PR #993: uf_finale_stage's new `git reset -q -- .uf/dewey/learnings` is a no-op in the repo that produced the PR (learnings tracked from the old policy) but runs with an unmatched pathspec in every scaffold consumer that never committed learnings — the exact population the mirrored plugin ships to — and whether git exits zero there is covered by no test in the suite (uf-finale tests mock exec; scratch-smoke checks tool names only; plugin-integration does not touch the finale plugin).
