---
tag: contract-testing
author: jay-flowers
category: pattern
created_at: 2026-11-25T06:36:52Z
identity: contract-testing-20261125T063652-jay-flowers
tier: draft
as_of: "unbound-force@e8eba4889b2e87b6dd4870ef340fc27ed7985c8e"
---

When a tool's failure codes must each be handled by every consuming command, pin the parity from both sides: the TypeScript suite iterates the tool's exported code constant against every command copy (canonical plus scaffold mirror) asserting each code appears in the prose, and a Go contract test pins the same clauses in both copies with a negative check that forbidden choreography (git checkout/switch/stash) never appears inside the gate section. Shipped as the review_head_root code-list parity test, TestCommandContracts_ReviewPrHeadRoot / TestCommandContracts_ReviewCouncilHeadRoot, and assertGatePathHasNoCheckout in internal/scaffold/command_contract_test.go. A newly added failure code that the command prose forgets to handle fails the TS parity suite even though the Go list is hardcoded, so the hardcoded Go list is redundancy rather than the sole guard.
