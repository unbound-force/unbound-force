---
tag: feedback-triage
author: jay-flowers
category: gotcha
created_at: 2026-11-05T15:40:17Z
identity: feedback-triage-20261105T154017-jay-flowers
tier: draft
---

Before accepting or re-confirming maintainer feedback on a PR, re-verify its factual premise at the frozen head. In PR #957 (address-feedback item 8), council feedback re-confirmed across two review rounds claimed the PR deleted TestPluginPackageInputs_RollbackPrerequisites and testdata/plugin-package-rollback.json; at head 60ca3d0 both exist (the test at scaffold_test.go:1825-1857 and the 11-line testdata file), the diff contains no such deletion, and the cited anchors "scaffold_test.go:1825/1923" are the retained test's own new/old line positions (cumulative hunk deltas -10, -95, +1, +6 map old line 1923 to new line 1825). Cheap checks settle presence: a file absent from the diff manifest is untouched by the PR, and a working-tree grep settles existence at head. Re-anchoring a prior finding's text without re-verification propagated a false premise into a "re-confirmed" review round; the likely source of confusion was the adjacent real deletion of a different lock-related test's content (TestPluginPackageInputs_ExactVersionsAndLockIntegrity's lock-assertion half and its assertLockedDirectPackage/assertLockedSecurityOverride helpers).
