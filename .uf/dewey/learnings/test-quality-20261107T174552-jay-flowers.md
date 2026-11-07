---
tag: test-quality
author: jay-flowers
category: pattern
created_at: 2026-11-07T17:45:52Z
identity: test-quality-20261107T174552-jay-flowers
tier: draft
---

Failure-path test suites for multi-branch migration helpers must be enumerated per branch, not per operation. In unbound-force PR #948, every merge-branch failure path of migrateSkillDir had a dedicated test (read, mkdir, move, remove — via chmod 0o555 and injected ReadFile errors), yet the rename-only branch's os.Rename failure went untested: the merge inventory never reaches it, and Run-level integration tests always scaffold skills/ before migrating, forcing the merge branch. Fix pattern: to force os.Rename failure deterministically in a Go test, chmod the rename's common parent directory to 0o555 and restore via t.Cleanup — Lstat/Stat preconditions still pass with r-x while Rename needs write on the parent and fails with EACCES (t.TempDir removal still succeeds because cleanups run LIFO with the restore registered later). Caveat: chmod-based failure injection is ineffective when tests run as root.
