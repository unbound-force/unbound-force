---
tag: scaffold
author: jay-flowers
category: pattern
created_at: 2026-11-07T19:55:28Z
identity: scaffold-20261107T195528-jay-flowers
tier: draft
---

When uf init gains a directory-shaping migration (e.g. migrateSkillDir renaming/merging legacy .opencode/skill/ into .opencode/skills/, PR 948 #933), the migration must be gated behind !opts.Stealth in Run() alongside the other directory migrations: stealth mode restores tracked .opencode/ paths after initSubTools, and moving tracked legacy files would leave the working tree modified, breaking stealth's clean-tree contract. A unit test on the migration function alone cannot catch a missing Run-level gate — the gate needs a two-mode Run test (normal mode migrates and prints the summary lines; stealth mode leaves both legacy directories byte-untouched), as TestRun_MigratesLegacyDirsExceptStealth does. This closes the class of round-4 finding #8 (stealth-gate change untested) for every future migration added to that block.
