---
tag: test-isolation
author: jay-flowers
category: gotcha
created_at: 2026-11-07T17:53:29Z
identity: test-isolation-20261107T175329-jay-flowers
tier: draft
---

Partial dependency injection in tests leaves the un-injected dependency falling through to its production default, which may touch the real host filesystem. In PR 948's unleash-delegate.test.ts, the bare-`..` traversal test injected fileExists (true for /fake) but omitted isDirectory, so the production default ran a real statSync("/fake"); on hosts where /fake exists as a directory, removing the traversal guard still passes the test because the fileExists && !isDirectory directory filter masks the removal — a silent pass-for-wrong-reason. Pattern: when a test's purpose is to detect removal of a guard, every dependency that shapes the guarded outcome must be injected or backed by a deterministic scratch tree, and the unguarded behavior must deterministically flip the assertion. Trap: pointing cwd at a mkdtemp directory does not close this class of gap — the parent (the system tmpdir) is a real directory, so the directory filter still masks removal; the ".."-resolved path must deterministically be a non-directory (inject isDirectory: () => false, or make it a real scratch file).
