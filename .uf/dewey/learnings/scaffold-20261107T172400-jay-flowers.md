---
tag: scaffold
author: jay-flowers
category: gotcha
created_at: 2026-11-07T17:24:00Z
identity: scaffold-20261107T172400-jay-flowers
tier: draft
---

When one guard block wraps multiple calls, the regression test must pin every guarded call, not just the newly added one. In PR #948, `if !opts.Stealth` (internal/scaffold/scaffold.go) wrapped both the new migrateSkillDir and the pre-existing migrateCommandDir, but TestRun_MigratesLegacySkillDirExceptStealth seeded only .opencode/skill/ fixtures — the command-dir half of the stealth no-mutation guarantee was unpinned, so a refactor re-enabling command-dir migration under stealth would pass the suite while reintroducing tracked-file mutation after restoreTrackedFiles. Pattern: author regression tests against the invariant the guard enforces (a stealth Run leaves the working tree unmodified) and seed fixtures for every call the guard protects; a half-pinned invariant lets a fixed defect silently return.
