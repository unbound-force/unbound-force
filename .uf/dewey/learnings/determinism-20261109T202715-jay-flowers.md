---
tag: determinism
author: jay-flowers
category: pattern
created_at: 2026-11-09T20:27:15Z
identity: determinism-20261109T202715-jay-flowers
tier: draft
---

When extracting command prose to a plugin tool that returns per-code failure envelopes, pin the tool's exported failure-code constant to the command's documented handling list with a parity test that reads every command copy (canonical and scaffold mirror) and requires exact equality. The test extracts the `- `CODE`:` bullets from the consuming command section and compares them against the exported code array, so a code added to the tool without command handling — or a handling line documenting a code the tool cannot return — fails the suite. Shipped in the checkout-pr-head change (issue #965): .opencode/test/checkout-pr-head.test.ts extracts the Section 1.2.1 bullets from both uf.address-feedback.md copies and asserts equality with CHECKOUT_PR_HEAD_ERROR_CODES.
