---
tag: change-records
author: jay-flowers
category: gotcha
created_at: 2026-10-31T19:52:39Z
identity: change-records-20261031T195239-jay-flowers
tier: draft
---

PR-body and commit-message claims must be verified against the cumulative base..head diff before approval. In the opsx/reviewer-capabilities-init-survival review (PR #890, 2026-10-31), the PR body and head commit message claimed a CLI asset-count test-range correction ([71-75] -> [73-77] for the 78-asset inventory) that appears nowhere in the PR diff — no matching numeric strings in any changed file, no cmd/ changes, and the scaffold_test.go diff carried exactly two hunks. Diff-anchored review cannot catch record claims about absent work, because the reviewer only examines changed lines. Cheap pre-merge check: for each claim in the PR body or commit message, grep its identifying tokens (file paths, numeric ranges, flag names) against the diff; any claim with no matching hunk is a record defect to resolve before merge — either include the work or amend the record — because post-merge the false claim becomes permanent history narrative.
