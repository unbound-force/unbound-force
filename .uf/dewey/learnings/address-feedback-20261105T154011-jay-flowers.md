---
tag: address-feedback
author: jay-flowers
category: gotcha
created_at: 2026-11-05T15:40:11Z
identity: address-feedback-20261105T154011-jay-flowers
tier: draft
---

Verify a re-confirmed maintainer feedback item's factual premise against the frozen head before accepting it — "re-confirmed from the prior review" is not evidence of re-verification. PR #957 item 8 claimed the PR deleted TestPluginPackageInputs_RollbackPrerequisites and its testdata while design.md:58 kept naming the test as a single-commit rollback safety net. At head 60ca3d0 the test exists at internal/scaffold/scaffold_test.go:1825, internal/scaffold/testdata/plugin-package-rollback.json exists, and design.md D5 explicitly retains both as a historical record that "nothing deploys or verifies". The feedback's own line citation pointed at the live test. A grep of each cited identifier against the head tree settles a DATA-DRIVEN classification in seconds; always grep artifacts a feedback item claims were deleted, because a v2 re-confirmation can carry a v1 premise past the revision that addressed it, and adjacent diff deletions (here, the deleted TestPluginPackageInputs_ExactVersionsAndLockIntegrity lock-assertion block and its helpers) are easily misread as the cited test's deletion.
