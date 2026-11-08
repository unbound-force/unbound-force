---
tag: address-feedback
author: jay-flowers
category: gotcha
created_at: 2026-11-05T15:50:20Z
identity: address-feedback-20261105T155020-jay-flowers
tier: draft
---

When validating maintainer feedback that claims a PR deletes a named test or fixture, verify the deletion claim against the complete per-file diff and the working tree at head before accepting it. PR 957 feedback item 8 claimed TestPluginPackageInputs_RollbackPrerequisites and testdata/plugin-package-rollback.json were deleted; the complete scaffold_test.go diff showed the named test in no hunk (retained verbatim at line 1825) and the testdata was never in the changed-files manifest. The reviewer had conflated neighboring deletions (the lock-assertion half of TestPluginPackageInputs_ExactVersionsAndLockIntegrity and its assertLocked* helpers). Review citation line numbers can point at retained code — a cited line number is evidence of where the reviewer looked, not evidence that the code changed.
