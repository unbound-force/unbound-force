---
tag: go-testing
author: jay-flowers
category: gotcha
created_at: 2026-11-07T19:55:43Z
identity: go-testing-20261107T195543-jay-flowers
tier: draft
---

chmod-0o555 failure injection in Go tests (forcing os.Rename/os.Remove to fail) stops working when the suite runs as root, because root bypasses permission bits — the mutation succeeds and the test fails loudly on the wrong assertion instead of exercising the failure path. When the syscall is not injectable through an Options seam (as ReadFile/WriteFile are in internal/scaffold), either gate the subtest on os.Geteuid() != 0 with t.Skip, or add the syscall wrapper to the injectable Options so failure paths are hermetic. Observed in PR #948's migrate_skill_test.go rename/move/remove failure tests.
