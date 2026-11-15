---
tag: preflight-baseline
author: jay-flowers
category: gotcha
created_at: 2026-11-15T20:47:47Z
identity: preflight-baseline-20261115T204747-jay-flowers
tier: draft
---

In unbound-force's local environment (observed 2026-11-15 on branch quick/extend-heredoc-ban-contract-test-scan-to), `make check` fails at the lint step even on code byte-identical to upstream/main: the locally installed golangci-lint (mise go 1.26.5 toolchain) reports an errcheck violation at internal/gateway/gateway_test.go:4383 that CI's linters do not flag (upstream/main CI is fully green). When a /uf.unleash or review pre-flight gate hits this, classify it as pre-existing by running the same local golangci-lint in a detached worktree at the merge-base (`git worktree add /tmp/opencode/preflight-baseline <merge-base> --detach`) — the identical finding reproduces there — and treat the gate as green over it; never enter the build/test fix loop for a failure that exists at the baseline. Also expect the documented local lock race during make plugin-test: a concurrent opencode host writes a gitignored .opencode/package-lock.json mid-run and trips the no-lock-shipped assertion; remove the stray lock and re-run the target, never delete node_modules.
