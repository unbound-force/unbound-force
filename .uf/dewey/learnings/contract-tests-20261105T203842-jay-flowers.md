---
tag: contract-tests
author: jay-flowers
category: pattern
created_at: 2026-11-05T20:38:42Z
identity: contract-tests-20261105T203842-jay-flowers
tier: draft
---

Command contract tests in internal/scaffold/command_contract_test.go use presence-only clause tables (assertCommandCopies -> strings.Contains on normalized text). Presence cannot pin ordering invariants: when a command fix's core contract is that one instruction precedes another — e.g., /uf.address-feedback Section 1.2.1 (PR-head checkout, line 146) must precede the first **CHECKPOINT** marker (line 293) so progress state keys to the PR branch (#954, PR #955) — the gate must also assert index ordering, following the file's existing structural precedents (count pins, forbidden fragments, regex boundaries in TestCommandContracts_FixCiCommandStructure). Anchor the ordering assertion on a marker unique to the guarded region: in uf.address-feedback.md the first get_progress mention (SESSION-RESUME GUARD, line 18) legitimately precedes Section 1.2.1, so anchor on the first **CHECKPOINT** marker or the section heading, not on get_progress. Iterate loadCommandCopies(t, name, true) so one assertion covers the canonical file and the byte-identical scaffold mirror.
