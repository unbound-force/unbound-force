---
tag: issue-triage
author: jay-flowers
category: gotcha
created_at: 2026-11-02T03:16:22Z
identity: issue-triage-20261102T031622-jay-flowers
tier: draft
---

When triaging gh-backed plugin-tool issues, map the issue's stated test mechanism to the repo's actual harness before accepting the criterion as written. Issue #669 (unbound-force) asks for "Tests added with mock BunShell", but the uf plugins deliberately use Bun.spawn argument arrays — NOT BunShell — for shell-injection safety (documented in .opencode/plugins/review-dispatch/index.ts runGhCommand), and the established mock is the shared Bun.spawn harness at .opencode/test/helpers/gh-spawn-mock.ts (used by the fetch-pr-review-context, fetch-pr-checks, and fetch-failed-run-logs test suites). The terminology imprecision is non-blocking because the intent (mock-based gh tests, no live network) is unambiguous; the correct triage action is to note the mapping so implementers reuse the existing harness instead of introducing a BunShell mock. Reinforcing why the test criterion is a hard gate rather than polish: .opencode/vitest.config.ts enforces a governance-controlled 94 percent statements / 85 percent branches coverage floor over plugins/** and lib/**, so any new uf-workflow tool shipped without mock-based tests fails make plugin-test mechanically — tool-extraction issues under Constitution Principle VI / DE-001 that omit the test criterion are incompletable as written.
