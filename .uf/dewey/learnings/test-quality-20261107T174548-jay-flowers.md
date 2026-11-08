---
tag: test-quality
author: jay-flowers
category: pattern
created_at: 2026-11-07T17:45:48Z
identity: test-quality-20261107T174548-jay-flowers
tier: draft
---

Deterministic filesystem-failure tests in Go: to exercise os.Rename/os.Remove error paths without root, chmod the directory that must be modified to 0o555 immediately before the operation and restore 0o755 via t.Cleanup so t.TempDir teardown still succeeds (shipped pattern: internal/scaffold/migrate_skill_test.go MoveFailure/RemoveFailure, lines 192/195 and 241/244). For a top-level directory rename, chmod the shared parent (e.g. .opencode), since rename needs write permission on the parent of both endpoints; all earlier steps in the path (Lstat/Stat) are reads and still succeed under 0o555, so the rename is the first failing syscall. Pair every new error-reporting branch with a matching XxxFailure test in the same increment — the suite's branch-to-test symmetry is the invariant that catches reporting regressions. Caveat: chmod-based denial does not bind uid 0, so these tests assume non-root execution (CI standard).
