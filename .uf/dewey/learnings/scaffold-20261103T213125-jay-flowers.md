---
tag: scaffold
author: jay-flowers
category: pattern
created_at: 2026-11-03T21:31:25Z
identity: scaffold-20261103T213125-jay-flowers
tier: draft
---

uf init's Run() restores tracked files for stealth mode (restoreTrackedFiles) BEFORE the legacy-directory migration block, so any new tree-mutating migration added after that point MUST be wrapped in `!opts.Stealth` — an unguarded migration would move tracked legacy files after the restore, dirtying the working tree and breaking the stealth cleanliness invariant that emitStealthSummary asserts. PR #948's migrateSkillDir followed this pattern and extended the guard to the pre-existing migrateCommandDir, fixi
