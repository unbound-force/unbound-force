---
tag: go-testing
author: jay-flowers
category: gotcha
created_at: 2026-11-07T17:45:50Z
identity: go-testing-20261107T174550-jay-flowers
tier: draft
---

Permission-based failure injection in Go tests — chmod 0o555 on a directory to force os.Rename/os.Remove to fail — is deterministic only for non-root test processes: root bypasses directory permission checks via CAP_DAC_OVERRIDE, the operation succeeds, and the test observes unexpected success. The internal/scaffold migration suite relies on this pattern (TestMigrateSkillDir_MoveFailure, TestMigrateSkillDir_RemoveFailure, and the rename-failure test recommended by PR 948 feedback thread PRRT_kwDORVoNbc604LsD). Run the suite as non-root, matching GitHub Actions runners; a failure of these tests inside a root container is environmental, not a regression. Always restore permissions via t.Cleanup so t.TempDir removal succeeds.
