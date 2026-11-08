---
tag: test-quality
author: jay-flowers
category: pattern
created_at: 2026-11-07T17:24:02Z
identity: test-quality-20261107T172402-jay-flowers
tier: draft
---

When one guard clause changes the behavior of multiple call sites (PR #948 wrapped both migrateCommandDir and migrateSkillDir in `if !opts.Stealth`), the regression test must pin EACH branch's gated behavior, not just one: a test that fails on wholesale gate removal still passes when a single branch is asymmetrically re-enabled out of the gate. In PR #948, TestRun_MigratesLegacySkillDirExceptStealth pinned the skill-dir half while the command-dir half — the half that changed pre-existing unconditional behavior — could be re-enabled under stealth without failing any test, because every other .opencode/command/ fixture lived in direct-call unit tests that bypass the Run()-level gate. Review rule for gate changes: enumerate one fixture+assertion pair per gated call site, and verify where each branch's existing fixtures live — direct unit tests of a function do not pin the caller's gate around it.
