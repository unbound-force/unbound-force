---
tag: test-quality
author: jay-flowers
category: pattern
created_at: 2026-11-07T16:02:11Z
identity: test-quality-20261107T160211-jay-flowers
tier: draft
---

When a diff moves a pre-existing operation inside a newly added conditional gate (e.g., wrapping both migrateCommandDir and migrateSkillDir in `if !opts.Stealth` in internal/scaffold/scaffold.go), the gate changes the pre-existing operation's contract in the gated state — and a gate-state test written for the newly added operation does not pin it. Each operation inside the gate needs its own gate-state test. Observed on PR #948: TestRun_MigratesLegacySkillDirExceptStealth covered the new skill-dir migration under stealth but not the command-dir migration whose stealth behavior changed in the same edit; a one-line extension of the stealth subtest (create .opencode/command/legacy.md, assert it survives) closes the gap.
