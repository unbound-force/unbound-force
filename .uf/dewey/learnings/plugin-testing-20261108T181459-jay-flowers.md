---
tag: plugin-testing
author: jay-flowers
category: pattern
created_at: 2026-11-08T18:14:59Z
identity: plugin-testing-20261108T181459-jay-flowers
tier: draft
---

Mocked argv assertions verify what a plugin tool says to git, not what git does. When a change adds a new git invocation to a tool's executor path (e.g. uf_finale_stage's `reset -q -- <dir>` in PR 993), mocked GitExec and mockGit tests stay green even if the real command fails or behaves differently on an unmatched pathspec. The unbound-force plugin suite has a DE-007 Layer 3 pattern for this: real-git scratch-repo round-trip tests that invoke the tool through the real executor and assert both the tool envelope and the resulting index/worktree state — precedents: preflight-declared-roundtrip.test.ts (cites DE-007 by name), review-diff.test.ts ("review-diff-real-" scratch repos), uf-workflow.test.ts ("uf-push-real-" bare-remote repos). Caveat: a module-level vi.mock("node:child_process") in a suite file blocks real git for the whole file — a round-trip test there needs vi.importActual, or it belongs in a dedicated round-trip test file.
