---
tag: preflight-baseline
author: jay-flowers
category: gotcha
created_at: 2026-11-16T03:57:06Z
identity: preflight-baseline-20261116T035706-jay-flowers
tier: draft
---

At unbound-force/unbound-force commit b6127cc451ac855bb8b3f198e3574a4ed04667ea (upstream/main, 2026-11-15), two checks are red on main itself and will surface in PR CI for any branch off that commit: golangci-lint errcheck "Error return value is not checked" at internal/gateway/gateway_test.go:4383, and cmd/unbound-force TestRunInit_FreshDir failing its file-count window (expected [84-88], got 89 — an uf init asset was added without bumping the window). The deterministic way to prove such failures pre-existing during a /uf.unleash pre-flight gate: git worktree add /tmp/opencode/preflight-baseline-<sha> <merge-base> --detach, run the exact failing check inside the worktree, then git worktree remove --force — the working tree is never touched, and an identical failure at the merge-base classifies as pre-existing (never enters the build/test fix loop; recorded in the gate's checkpoint inventory instead). Also on that machine: make plugin-test can trip its no-lock-shipped assertion from the documented local lock race (concurrent opencode hosts writing a gitignored .opencode/package-lock.json mid-run); the remedy is remove the stray lock and re-run, never delete node_modules. Provenance: branch quick/pin-the-triage-issue-posting-choreograph, /uf.unleash quick-tier pre-flight hard gate, 2026-11-15.
