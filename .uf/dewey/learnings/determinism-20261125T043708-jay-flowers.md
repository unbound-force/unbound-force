---
tag: determinism
author: jay-flowers
category: gotcha
created_at: 2026-11-25T04:37:08Z
identity: determinism-20261125T043708-jay-flowers
tier: draft
as_of: "unbound-force@15b2ed01d5d1d00d9cc6797040a2d55fcd01f02b"
---

An extraction-fix scope must include every exact-shape pin the original change added, not just the golden output-shape suite: the unit suite's toEqual on the tool's full output shape (review-head-root.test.ts:234), the Go command-contract fragments that pin the reworded prose (command_contract_test.go clauses matching e.g. `local gates: not run (<code>): <message>`), and the CHANGELOG DE-004 deltas — which go stale whenever later commits add lines to the same instruction files (PR #1211 recorded 599→634 / 1309→1349 / 1487→1538 while the merge-head linter delta reports 637 / 1353 / 1543). All of these fail closed. Re-run `instruction-lint --json --base <ref> --changed` at fix time and paste the terminal numbers rather than the numbers captured when the entry was drafted.
