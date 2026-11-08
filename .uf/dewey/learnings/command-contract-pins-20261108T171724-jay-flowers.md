---
tag: command-contract-pins
author: jay-flowers
category: pattern
created_at: 2026-11-08T17:17:24Z
identity: command-contract-pins-20261108T171724-jay-flowers
tier: draft
---

When a command-prose change adds MUST behavioral rules to uf.review-pr.md or uf.address-feedback.md, extend the behavioral contract pins in the same PR, not just the option-label pins. The repo maintains two pin layers for these commands: assertCommandCopies clause tests in internal/scaffold/command_contract_test.go (precedent: TestCommandContracts_ReviewPrSelfAuthoredMapping pinned issue #942's verdict-mapping rule when it landed) and assertFragments tests in internal/scaffold/scaffold_test.go (including 'legacy fragments absent' guards and the 'bounded mechanical retry policy' pin). Option-label/option-array pins in TestQuestionToolPresentationContractGates lock structure, not values — thresholds (e.g. a 200-line delta cap), eligibility condition lists, sensitive-path lists, attempt caps, and removed-gate absence all need fragment pins to be regression-protected; one loadCommandCopies assertion covers canonical and embedded copies via the drift-enforced mirror.
