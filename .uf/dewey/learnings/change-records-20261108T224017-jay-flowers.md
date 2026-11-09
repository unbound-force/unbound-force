---
tag: change-records
author: jay-flowers
category: gotcha
created_at: 2026-11-08T22:40:17Z
identity: change-records-20261108T224017-jay-flowers
tier: draft
---

Instruction-size (DE-004) measurements recorded in change records go stale when a PR is rebased onto a newer main: numbers computed against the pre-rebase merge-base (PR #959's linguist "236 -> 203" was measured before PR #958's frontmatter hardening grew the base file to 262) no longer match the final base-to-head diff, and new-file counts drift when post-measurement commits grow the file (determinism recorded 246, shipped 263). Before merging any change that publishes line/token deltas, re-run the deltas tool at the final head (`.uf/scripts/instruction-lint --json --base origin/main --changed`) and paste its output — the record's evidence must be measured against the merge-base the PR will actually merge through, and verify each Docs claim in the proposal/tasks against the final diff (PR #959 claimed an AGENTS.md agent-count edit that no commit carried).
