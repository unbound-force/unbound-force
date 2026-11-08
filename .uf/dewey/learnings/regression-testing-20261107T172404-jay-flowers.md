---
tag: regression-testing
author: jay-flowers
category: pattern
created_at: 2026-11-07T17:24:04Z
identity: regression-testing-20261107T172404-jay-flowers
tier: draft
---

When a diff wraps multiple call sites behind a new conditional gate, the regression test must pin every gated site — especially pre-existing call sites whose behavior changes, which are easy to skip because the test gets written for the new code. Precedent: PR 948 gated both migrateCommandDir (pre-existing, previously unconditional) and migrateSkillDir (new) behind !opts.Stealth in scaffold Run(); TestRun_MigratesLegacySkillDirExceptStealth pinned only the skill-dir half, leaving the command-dir behavior change unpinned until maintainer feedback. Pattern: enumerate the call sites inside a new gate and give each one a fixture plus a survival/mutation assertion in the gating test.
