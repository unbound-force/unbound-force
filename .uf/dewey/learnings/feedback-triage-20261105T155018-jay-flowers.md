---
tag: feedback-triage
author: jay-flowers
category: gotcha
created_at: 2026-11-05T15:50:18Z
identity: feedback-triage-20261105T155018-jay-flowers
tier: draft
---

When triaging review feedback that claims a PR deletes a named test, verify the claim against the diff's actual deletion set before accepting it — and treat the feedback's own line citations as a lead, not proof. Verified on PR 957 feedback item 8 (unbound-force/unbound-force): the feedback claimed design.md D5 referenced a deleted test and cited scaffold_test.go:1825/1923, but line 1825 is where TestPluginPackageInputs_RollbackPrerequisites lives at head; the PR's deletion set removed only the lock-assertion half of a differently-named test (TestPluginPackageInputs_ExactVersionsAndLockIntegrity, renamed TestPluginPackageInputs_ExactVersions) plus its assertLocked* helpers. Renamed or trimmed tests in a diff are especially prone to this conflation — a reviewer scanning deletions nearby a retained test can misattribute the deletion. Cheap verification sequence: (1) grep the named test and its testdata in the working tree at head, (2) read the file's full diff chunk to enumerate the actual deletion set, (3) read the design/spec paragraph the feedback characterizes to check the characterization, not just the file reference. Complements issue-triage learnings on verifying cited artifacts against current repo state; adds the review-feedback variant where the feedback's line numbers often point at live code, not deleted code.
