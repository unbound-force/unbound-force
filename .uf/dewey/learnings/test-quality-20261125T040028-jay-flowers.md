---
tag: test-quality
author: jay-flowers
category: pattern
created_at: 2026-11-25T04:00:28Z
identity: test-quality-20261125T040028-jay-flowers
tier: draft
as_of: "unbound-force@15b2ed01d5d1d00d9cc6797040a2d55fcd01f02b"
---

Deriving documentation-parity tests from a plugin tool's exported failure-code const (rather than a list hardcoded in the test) turns adding a code into a forcing function: review-head-root.test.ts's code-list parity describe iterates REVIEW_HEAD_ROOT_ERROR_CODES and fails against every consuming command copy (canonical .opencode/commands/ + internal/scaffold/assets mirror) until each code appears in the text, so the TS test surface stays in lockstep with the tool's documented contract with zero list duplication. The Go command_contract_test.go keeps a hardcoded prepare-code list instead (cross-language derivation is impractical), with a comment naming review-head-root.test.ts as the TS parity counterpart — an acceptable split as long as the comment records the relationship. When a code must be skipped in such a loop (INVALID_ROOT, handled as a generic release warning rather than a prepare-failure code), an inline comment stating the compensation (the skill contract test pins the Phase 2b INVALID_ROOT handling) keeps the skip auditable instead of looking like an oversight.
