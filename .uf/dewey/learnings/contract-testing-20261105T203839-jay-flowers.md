---
tag: contract-testing
author: jay-flowers
category: pattern
created_at: 2026-11-05T20:38:39Z
identity: contract-testing-20261105T203839-jay-flowers
tier: draft
---

Contract tests that pin command content by substring presence cannot detect reordering regressions — presence is order-insensitive. When a fix's invariant is positional (step A must precede step B, e.g. a PR-head checkout before the first checkpoint/get_progress call so progress state is keyed to the right branch), the regression test must assert index ordering (strings.Index(content, anchorA) < strings.Index(content, anchorB)) and fail closed (t.Fatalf) when either anchor is missing. Anchor choice matters: pick the marker whose FIRST occurrence is the normative one. Conditional prose that merely mentions an instruction (a session-resume guard saying "call get_progress") can appear earlier in the file than the step it must follow, so anchoring on the instruction's first mention can fail against a correct file; a structural marker that exists only at the gated sites (the first **CHECKPOINT** block) is the safe anchor. Precedent: internal/scaffold/command_contract_test.go — FixCi's countPins and forbidden-fragment loops and SpeckitTestReview's post-helper structural check already go beyond presence for exactly this reason (found assessing PR #955 feedback item 8, issue #954).
