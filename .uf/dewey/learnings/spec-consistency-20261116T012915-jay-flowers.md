---
tag: spec-consistency
author: jay-flowers
category: gotcha
created_at: 2026-11-16T01:29:15Z
identity: spec-consistency-20261116T012915-jay-flowers
tier: draft
as_of: "unbound-force@1677fd12d4ed8969d5ac87110e35ca030b9326b1"
evidence:
  - "gh:unbound-force/unbound-force#1048"
  - "dewey:.uf/artifacts/dispatch/a504842e-57da-41a4-926a-884c37671964.json"
---

Cross-artifact regex drift in openspec/changes/heredoc-ban-escaped-and-ansic: design.md D1 and tasks.md 1.1 record the target heredoc-ban pattern as `<<-?\s*\$?\\?['\"]?[A-Za-z_][A-Za-z0-9_]*['\"]?` (no whitespace between the optional backslash and the optional quote; D1's own enumeration lists only "-, whitespace, $, \, quote, identifier"), while the spec delta SC-FR-016's canonical spelling and both compiled copies in internal/scaffold/command_contract_test.go use `<<-?\s*\$?\\?\s*['\"]?[A-Za-z_][A-Za-z0-9_]*['\"]?`. The inner `\s*` between `\\?` and `['\"]?` is load-bearing: it is the only thing that matches the pinned positive `<<\ 'EOF'`, so an implementer following tasks.md 1.1 verbatim produces a regex that fails the table-driven subtest task 1.2 adds (verified by regex analysis: without the inner `\s*`, `<<\ 'EOF'` cannot match at any position). When a spec delta declares a canonical pattern spelling, design decisions and task steps should reference it ("the canonical spelling in SC-FR-016") rather than restating it — every restatement is a drift surface where one dropped token silently contradicts the change's own acceptance tests.
