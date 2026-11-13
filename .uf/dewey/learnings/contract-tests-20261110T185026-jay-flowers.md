---
tag: contract-tests
author: jay-flowers
category: pattern
created_at: 2026-11-10T18:50:26Z
identity: contract-tests-20261110T185026-jay-flowers
tier: draft
---

Scan-based rejection tests (asserting a forbidden pattern is absent from real repo assets) run only their no-violation path in a green suite, so the detector itself is never exercised: the fail-without-fix evidence lives only in VCS history. Pin the detector's match set with a companion table-driven positive control — hoist the pattern to a shared test-level var and assert the match cases (quoted/unquoted/spaced/dashed delimiter forms) plus documented no-match edge limitations — otherwise a future narrowing of the regex silently disables the guard while every check stays green. Observed on TestCommandContracts_RejectShellHeredocPayloads (#1011 heredoc guard, internal/scaffold/command_contract_test.go): the regex matches the removed `<< 'COMMENT_EOF'` lines, so the regression property holds, but nothing in the committed suite would catch a regex regression.
