---
tag: plugin-testing
author: jay-flowers
category: pattern
created_at: 2026-11-09T20:27:32Z
identity: plugin-testing-20261109T202732-jay-flowers
tier: draft
---

When testing an OpenCode plugin tool that wraps git/gh choreography, inject executor fakes over a mutable repository model and give the gh fake an onCheckout hook that mutates the model (moves HEAD, renames the branch, rewrites branch config) — the tool's post-checkout re-verification then runs against a real state transition instead of a canned return value. Record every invocation as an argument array and assert exact arrays (toEqual on [["pr","checkout","42"]]) to pin no-shell execution, exact call counts (two rev-parse HEAD calls) to pin re-verify steps, and negative call assertions (no merge-base/status after an equality short-circuit; no config reads before the branch charset gate) to pin decision ordering. Shipped in .opencode/test/checkout-pr-head.test.ts (27 tests, PR #1002).
