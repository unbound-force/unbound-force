---
tag: injection-safety
author: jay-flowers
category: context
created_at: 2026-11-16T02:24:02Z
identity: injection-safety-20261116T022402-jay-flowers
tier: draft
as_of: "unbound-force@1677fd12d4ed8969d5ac87110e35ca030b9326b1"
evidence:
  - "dewey:.uf/artifacts/dispatch/fb460712-fb0f-49e7-bf32-4362f5c4eba8.json"
  - "dewey:internal/scaffold/command_contract_test.go"
  - "learning/injection-safety-20261110T234758-jay-flowers"
  - "session:ses_current-uf-pipeline-1048"
---

The heredoc-ban regex gap flagged in learning/injection-safety-20261110T234758-jay-flowers is closed by openspec/changes/heredoc-ban-escaped-and-ansic (#1048): TestCommandContracts_RejectShellHeredocPayloads now matches escaped-delimiter (<<\EOF, <<\ 'EOF') and ANSI-C-quoted (<<$'EOF') heredocs, the recognized-spelling set is pinned by TestCommandContracts_HeredocPattern with count guards (15 positives, 4 negatives, 19 total), and TestCommandContracts_HeredocScanFixture pins the walk-to-message path via a t.TempDir fixture. Reusable pattern: when a ban regex's coverage is the contract, assert the table's row counts alongside the cases, so deleting a row fails the count check instead of silently narrowing the ban.
