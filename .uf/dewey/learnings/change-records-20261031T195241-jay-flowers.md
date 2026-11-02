---
tag: change-records
author: jay-flowers
category: gotcha
created_at: 2026-10-31T19:52:41Z
identity: change-records-20261031T195241-jay-flowers
tier: draft
---

PR bodies and commit messages must be diff-verified before submission: every claimed change should be checkable against an actual hunk in the shipped diff. When commit-scope rules split tooling/scaffold refreshes onto a separate chore branch (e.g. chore/uf.init-sync), the feature PR's change record must replace the moved claim with an explicit cross-reference to the branch or PR that carries it, rather than retaining the claim verbatim — otherwise release notes drafted from the record (a planned follow-up in the change's docs issue) describe changes the PR does not carry. Caught in the PR #890 review (reviewer-capabilities-init-survival): the PR body and head commit message claimed a "[71-75] -> [73-77]" CLI asset-count test-range correction for a 78-asset inventory that appears nowhere in the diff — scaffold_test.go carried only the regression-test and TestIsToolOwned table-flip hunks, and no cmd/ or asset files changed.
